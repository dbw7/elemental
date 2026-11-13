/*
Copyright © 2025-2026 SUSE LLC
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package archive

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/suse/elemental/v3/pkg/sys"
	"github.com/suse/elemental/v3/pkg/sys/vfs"
)

type Filter func(h *tar.Header) (bool, error)

type cancelableReader struct {
	ctx context.Context
	src io.Reader
}

func (r *cancelableReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, fmt.Errorf("stop reading, context cancelled")
	default:
		return r.src.Read(p)
	}
}

func newCancelableReader(ctx context.Context, src io.Reader) *cancelableReader {
	return &cancelableReader{
		ctx: ctx,
		src: src,
	}
}

type link struct {
	Name string
	Path string
}

func CreateTar(ctx context.Context, s *sys.System, sourceDir, target string) (err error) {
	file, err := s.FS().OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, vfs.FilePerm)
	if err != nil {
		return fmt.Errorf("creating archive %s: %w", target, err)
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	tw := tar.NewWriter(file)
	defer func() {
		err = errors.Join(err, tw.Close())
	}()

	return vfs.WalkDirFs(s.FS(), sourceDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == sourceDir {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("stop writing tar, context cancelled")
		default:
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		var linkTarget string
		if info.Mode()&os.ModeSymlink != 0 {
			if linkTarget, err = s.FS().Readlink(path); err != nil {
				return fmt.Errorf("reading symlink %s: %w", path, err)
			}
		}

		header, err := tar.FileInfoHeader(info, linkTarget)
		if err != nil {
			return fmt.Errorf("creating tar header for %s: %w", path, err)
		}

		if header.Name, err = filepath.Rel(sourceDir, path); err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		if err = tw.WriteHeader(header); err != nil {
			return fmt.Errorf("writing tar header for %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		src, err := s.FS().Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		if _, err = io.Copy(tw, newCancelableReader(ctx, src)); err != nil {
			return fmt.Errorf("writing %s to tar: %w", path, err)
		}

		return nil
	})
}

// ExtractTarball extracts a .tar, .tar.gz or .tar.bz2 taball file to the given target.
// Compression detection is rudimentary and only based on file name extension.
func ExtractTarball(ctx context.Context, s *sys.System, tarball string, target string, filters ...Filter) error {
	sourceFile, err := s.FS().OpenFile(tarball, os.O_RDONLY, vfs.FilePerm)
	if err != nil {
		return err
	}
	switch {
	case strings.HasSuffix(tarball, "tar.bz2"):
		return ExtractTarBz2(ctx, s, sourceFile, target, filters...)
	case strings.HasSuffix(tarball, "tar.gz"):
		return ExtractTarGz(ctx, s, sourceFile, target, filters...)
	default:
		return ExtractTar(ctx, s, sourceFile, target, filters...)
	}
}

// ExtractTarGz extracts a ..tar.gz archived stream of data to the given target
func ExtractTarGz(ctx context.Context, s *sys.System, body io.Reader, target string, filters ...Filter) error {
	reader, err := gzip.NewReader(body)
	if err != nil {
		return fmt.Errorf("gzip error: %w", err)
	}

	return ExtractTar(ctx, s, reader, target, filters...)
}

// ExtractTarBz2 extracts a ..tar.bz2 archived stream of data to the given target
func ExtractTarBz2(ctx context.Context, s *sys.System, body io.Reader, target string, filters ...Filter) error {
	reader := bzip2.NewReader(body)
	return ExtractTar(ctx, s, reader, target, filters...)
}

// ExtractTar extracts a .tar archived stream of data to the given target
func ExtractTar(ctx context.Context, s *sys.System, body io.Reader, target string, filters ...Filter) error {
	var links []*link
	var symlinks []*link

	tr := tar.NewReader(body)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("stop reading tar, context cancelled")
		default:
		}

		header, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else if !errors.Is(err, tar.ErrInsecurePath) {
				return fmt.Errorf("reading tar stream: %w", err)
			}

			s.Logger().Warn("Ignoring non local path '%s': %v", header.Name, err)
			continue
		}

		path, err := filterTarHeader(s, target, header, filters...)
		if err != nil {
			return fmt.Errorf("tar filter failed: %w", err)
		}
		if path == "" {
			continue
		}

		info := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeDir:
			if err = vfs.MkdirAll(s.FS(), path, info.Mode()); err != nil {
				return fmt.Errorf("creating directory from tar: %w", err)
			}
		case tar.TypeReg:
			if err = copyFile(ctx, s, path, info.Mode(), tr); err != nil {
				return fmt.Errorf("creating file %s: %w", path, err)
			}
		case tar.TypeLink:
			name := header.Linkname
			name, err = sanitizeArchivePath(target, name)
			if err != nil {
				s.Logger().Warn("Ignoring non local path '%s': %v", name, err)
				continue
			}
			links = append(links, &link{Path: path, Name: name})
		case tar.TypeSymlink:
			symlinks = append(symlinks, &link{Path: path, Name: header.Linkname})
		}
	}

	for _, link := range links {
		_ = s.FS().Remove(link.Path)
		if err := s.FS().Link(link.Name, link.Path); err != nil {
			return fmt.Errorf("creating link %s: %w", link.Path, err)
		}
	}

	for _, symlink := range symlinks {
		_ = s.FS().Remove(symlink.Path)
		if err := s.FS().Symlink(symlink.Name, symlink.Path); err != nil {
			return fmt.Errorf("creating symlink %s: %w", symlink.Path, err)
		}
	}

	return nil
}

func copyFile(ctx context.Context, s *sys.System, path string, mode os.FileMode, src io.Reader) (err error) {
	dir := filepath.Dir(path)
	info, err := s.FS().Lstat(dir)
	switch {
	case os.IsNotExist(err):
		err = vfs.MkdirAll(s.FS(), dir, vfs.DirPerm)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	case info.Mode()&0200 == 0:
		// Ensure we can feed directories tared without write permission
		err = s.FS().Chmod(dir, vfs.DirPerm)
		if err != nil {
			return err
		}
		defer func() {
			e := s.FS().Chmod(dir, info.Mode())
			if err == nil && e != nil {
				err = e
			}
		}()
	}

	_ = s.FS().Remove(path)

	file, err := s.FS().OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer func() {
		e := file.Close()
		if err == nil && e != nil {
			err = e
		}
	}()
	_, err = io.Copy(file, newCancelableReader(ctx, src))
	return err
}

func sanitizeArchivePath(root, filename string) (string, error) {
	path := filepath.Join(root, filename)
	if strings.HasPrefix(path, filepath.Clean(root)) {
		return path, nil
	}

	return path, fmt.Errorf("content filepath '%s' is tainted", path)
}

func filterTarHeader(s *sys.System, target string, h *tar.Header, filters ...Filter) (string, error) {
	path := h.Name
	path, err := sanitizeArchivePath(target, path)
	if err != nil {
		s.Logger().Warn("Ignoring non local path '%s': %v", path, err)
		return "", nil
	}

	for _, filter := range filters {
		accept, err := filter(h)
		if err != nil {
			return "", fmt.Errorf("tar filter failed: %w", err)
		}
		if !accept {
			return "", nil
		}
	}

	return path, nil
}
