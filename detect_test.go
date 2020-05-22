package main_test

import (
	"errors"
	"testing"

	"github.com/paketo-buildpacks/packit"
	main "github.com/paketo-community/python-runtime"
	"github.com/paketo-community/python-runtime/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildpackYMLParser *fakes.VersionParser

		detect packit.DetectFunc
	)

	it.Before(func() {

		buildpackYMLParser = &fakes.VersionParser{}

		detect = main.Detect(buildpackYMLParser)
	})

	it("returns a plan that provides python", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: "python"},
			},
		}))

	})

	context("when the source code contains a buildpack.yml file", func() {
		it.Before(func() {
			buildpackYMLParser.ParseVersionCall.Returns.Version = "some-version"
		})

		it("returns a plan that provides and requires that version of python", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: main.Python},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name:    "python",
						Version: "some-version",
						Metadata: main.BuildPlanMetadata{
							VersionSource: "buildpack.yml",
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("when the buildpack.yml parser fails", func() {
			it.Before(func() {
				buildpackYMLParser.ParseVersionCall.Returns.Err = errors.New("failed to parse buildpack.yml")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "/working-dir",
				})
				Expect(err).To(MatchError("failed to parse buildpack.yml"))
			})
		})
	})
}
