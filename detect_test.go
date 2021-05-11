package cpython_test

import (
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit"
	cpython "github.com/paketo-community/cpython"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		detect packit.DetectFunc
	)

	it.Before(func() {

		detect = cpython.Detect()
	})

	it("returns a plan that provides cpython", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: "/working-dir",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: cpython.Cpython},
			},
		}))

	})

	context("when the BP_CPYTHON_VERSION env var is set", func() {
		it.Before(func() {
			os.Setenv("BP_CPYTHON_VERSION", "some-version")
		})

		it.After(func() {
			os.Unsetenv("BP_CPYTHON_VERSION")
		})

		it("returns a plan that provides and requires that version of cpython", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "/working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: cpython.Cpython},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: cpython.Cpython,
						Metadata: cpython.BuildPlanMetadata{
							Version:       "some-version",
							VersionSource: "BP_CPYTHON_VERSION",
						},
					},
				},
			}))
		})
	})
}
