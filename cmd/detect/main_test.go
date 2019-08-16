package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/python-cnb/python"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	when("testing versions", func() {
		when("there is no buildpack.yml", func() {
			it("shouldn't set the version in the buildplan", func() {
				runDetectAndExpectBuildplan(factory, buildplan.Plan{
					Provides: []buildplan.Provided{{Name: python.Dependency}},
				}, t)
			})
		})

		when("there is a buildpack.yml", func() {
			const version string = "1.2.3"

			it.Before(func() {
				buildpackYAMLString := fmt.Sprintf("python:\n  version: %s", version)
				Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), 0666, buildpackYAMLString)).To(Succeed())
			})

			it("should pass with the requested version of python", func() {
				runDetectAndExpectBuildplan(factory, buildplan.Plan{
					Provides: []buildplan.Provided{{Name: python.Dependency}},
					Requires: []buildplan.Required{{Name: python.Dependency,
						Version:  version,
						Metadata: buildplan.Metadata{"launch": true},
					}},
				}, t)
			})
		})
	})
}

func runDetectAndExpectBuildplan(factory *test.DetectFactory, buildplan buildplan.Plan, t *testing.T) {
	Expect := NewWithT(t).Expect

	code, err := runDetect(factory.Detect)
	Expect(err).NotTo(HaveOccurred())

	Expect(code).To(Equal(detect.PassStatusCode))

	Expect(factory.Plans.Plan).To(Equal(buildplan))
}
