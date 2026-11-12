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

package version

import (
	"fmt"
	"strings"
)

var (
	version = "v3.0"
	// gitCommit is the git sha1
	gitCommit = ""
)

func Version() string {
	commit := gitCommit

	if commit == "" {
		return version
	}

	if len(commit) > 7 {
		commit = gitCommit[:7]
	}

	// if the version already contains metadata just append
	// to the existing metadata with a dot
	if strings.Contains(version, "+") {
		return fmt.Sprintf("%s.g%s", version, commit)
	}

	return fmt.Sprintf("%s+g%s", version, commit)
}
