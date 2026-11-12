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

package semverutil

import (
	"github.com/Masterminds/semver/v3"
)

// Version represents a parsed semantic version.
type Version struct {
	v *semver.Version
}

// Parse parses a given version string into a Version object.
func Parse(version string) (Version, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return Version{}, err
	}
	return Version{v: v}, nil
}

// Major returns the major version.
func (v Version) Major() uint64 {
	if v.v == nil {
		return 0
	}
	return v.v.Major()
}

// Minor returns the minor version.
func (v Version) Minor() uint64 {
	if v.v == nil {
		return 0
	}
	return v.v.Minor()
}

// Patch returns the patch version.
func (v Version) Patch() uint64 {
	if v.v == nil {
		return 0
	}
	return v.v.Patch()
}

// Prerelease returns the pre-release portion, if any
func (v Version) Prerelease() string {
	if v.v == nil {
		return ""
	}
	return v.v.Prerelease()
}

// Metadata returns the metadata portion, if any
func (v Version) Metadata() string {
	if v.v == nil {
		return ""
	}
	return v.v.Metadata()
}

// String returns the string representation of the version.
func (v Version) String() string {
	if v.v == nil {
		return ""
	}
	return v.v.String()
}

func (v Version) LessThan(other Version) bool {
	if v.v == nil || other.v == nil {
		return false
	}
	return v.v.LessThan(other.v)
}

func (v Version) LessThanEqual(other Version) bool {
	if v.v == nil || other.v == nil {
		return false
	}
	return v.v.LessThanEqual(other.v)
}

func (v Version) GreaterThan(other Version) bool {
	if v.v == nil || other.v == nil {
		return false
	}
	return v.v.GreaterThan(other.v)
}

func (v Version) GreaterThanEqual(other Version) bool {
	if v.v == nil || other.v == nil {
		return false
	}
	return v.v.GreaterThanEqual(other.v)
}

func (v Version) Equal(other Version) bool {
	if v.v == nil || other.v == nil {
		return false
	}
	return v.v.Equal(other.v)
}
