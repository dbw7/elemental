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

package semverutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/suse/elemental/v3/pkg/semverutil"
)

func TestSemverUtilsSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Semver utilities test suite")
}

var _ = Describe("semver utilities", Label("semver"), func() {
	It("parses a valid semver version from string", func() {
		v, err := semverutil.Parse("3.1.0-rc2+g31.adb1438")
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Major()).To(Equal(uint64(3)))
		Expect(v.Minor()).To(Equal(uint64(1)))
		Expect(v.Patch()).To(Equal(uint64(0)))
		Expect(v.Metadata()).To(Equal("g31.adb1438"))
		Expect(v.Prerelease()).To(Equal("rc2"))
	})

	It("initates a version to zero", func() {
		v := semverutil.Version{}
		Expect(v.Major()).To(Equal(uint64(0)))
		Expect(v.Minor()).To(Equal(uint64(0)))
		Expect(v.Patch()).To(Equal(uint64(0)))
		Expect(v.Metadata()).To(BeEmpty())
		Expect(v.Prerelease()).To(BeEmpty())
	})

	It("errors out parsing invalid version", func() {
		_, err := semverutil.Parse("bad3.1.0-rc2+g31.adb1438")
		Expect(err).To(HaveOccurred())
	})

	It("compares semver versions", func() {
		v0, err := semverutil.Parse("v2.3")
		Expect(err).NotTo(HaveOccurred())
		v1, err := semverutil.Parse("2.3.0")
		Expect(err).NotTo(HaveOccurred())
		v2, err := semverutil.Parse("2.3.1-rc1")
		Expect(err).NotTo(HaveOccurred())
		v3, err := semverutil.Parse("2.3.1-rc2")
		Expect(err).NotTo(HaveOccurred())
		v4, err := semverutil.Parse("2.3.1+g31")
		Expect(err).NotTo(HaveOccurred())
		v5, err := semverutil.Parse("2.3.1+g45")
		Expect(err).NotTo(HaveOccurred())

		Expect(v1.Equal(v0)).To(BeTrue())

		By("using pre-release versions")
		Expect(v1.Equal(v2)).To(BeFalse())
		Expect(v2.GreaterThan(v1)).To(BeTrue())
		Expect(v3.GreaterThan(v2)).To(BeTrue())
		Expect(v3.LessThan(v4)).To(BeTrue())

		By("using metadata does not affect comparison")
		Expect(v5.GreaterThan(v4)).To(BeFalse())
		Expect(v5.GreaterThanEqual(v4)).To(BeTrue())
		Expect(v5.LessThanEqual(v4)).To(BeTrue())
		Expect(v5.Equal(v4)).To(BeTrue())

		By("empty version always results to false")
		Expect(v1.GreaterThan(semverutil.Version{})).To(BeFalse())
		Expect(v1.LessThan(semverutil.Version{})).To(BeFalse())
		Expect(v1.Equal(semverutil.Version{})).To(BeFalse())
		Expect(v1.GreaterThanEqual(semverutil.Version{})).To(BeFalse())
		Expect(v1.LessThanEqual(semverutil.Version{})).To(BeFalse())
	})
})
