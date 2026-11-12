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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVersionSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Version test suite")
}

var _ = Describe("Version tests", Label("version"), func() {
	It("returns the version without appending metadata if no commit is defined", func() {
		version = "v3.4.5"
		gitCommit = ""
		Expect(Version()).To(Equal("v3.4.5"))
	})
	It("returns the version adding a metadata field if a commit is defined", func() {
		version = "v3.4.5"
		gitCommit = "abed252"
		Expect(Version()).To(Equal("v3.4.5+gabed252"))
	})
	It("returns the version appending the commit to the existing metadata", func() {
		version = "v3.4.5-rc2+build3"
		gitCommit = "abed252"
		Expect(Version()).To(Equal("v3.4.5-rc2+build3.gabed252"))
	})
})
