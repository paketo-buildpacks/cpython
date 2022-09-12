package cpython_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/cpython"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		detect        packit.DetectFunc
		detectContext packit.DetectContext
	)

	it.Before(func() {
		detect = cpython.Detect()
		detectContext = packit.DetectContext{
			WorkingDir: "/working-dir",
		}
	})

	it("returns a plan that provides cpython", func() {
		result, err := detect(detectContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: cpython.Cpython},
			},
		}))

	})

	context("when the BP_CPYTHON_VERSION env var is set", func() {
		it.Before(func() {
			t.Setenv("BP_CPYTHON_VERSION", "some-version")
		})

		it("returns a plan that provides and requires that version of cpython", func() {
			result, err := detect(detectContext)
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

		it("returns a plan that provides and requires that version of cpython with configure flags", func() {
			t.Setenv("BP_CPYTHON_CONFIGURE_FLAGS", "--flag1 --flag2=value")
			result, err := detect(detectContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: cpython.Cpython},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: cpython.Cpython,
						Metadata: cpython.BuildPlanMetadata{
							Version:        "some-version",
							VersionSource:  "BP_CPYTHON_VERSION",
							ConfigureFlags: "--flag1 --flag2=value",
						},
					},
				},
			}))
		})
	})

	context("when there is a buildpack.yml", func() {
		var workingDir string
		it.Before(func() {
			var err error
			workingDir, err = os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(os.WriteFile(filepath.Join(workingDir, "buildpack.yml"), nil, os.ModePerm))

			detectContext.WorkingDir = workingDir
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("fails the build with a deprecation notice", func() {
			_, err := detect(detectContext)
			Expect(err).To(MatchError("working directory contains deprecated 'buildpack.yml'; use environment variables for configuration"))
		})
	})

	context("failure cases", func() {
		context("when there is an error determining if buildpack.yml exists", func() {
			var workingDir string

			it.Before(func() {
				var err error
				workingDir, err = os.MkdirTemp("", "working-dir")
				Expect(err).NotTo(HaveOccurred())

				Expect(os.Chmod(workingDir, 0000)).To(Succeed())

				detectContext.WorkingDir = workingDir
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
				Expect(os.RemoveAll(workingDir)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := detect(detectContext)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
