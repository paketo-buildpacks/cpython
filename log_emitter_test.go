package main_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	main "github.com/paketo-community/python-runtime"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer  *bytes.Buffer
		emitter main.LogEmitter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = main.NewLogEmitter(buffer)
	})

	context("Title", func() {
		it("logs the buildpack title", func() {
			emitter.Title(packit.BuildpackInfo{
				Name:    "some-name",
				Version: "some-version",
			})
			Expect(buffer.String()).To(Equal("some-name some-version\n"))
		})
	})

	context("Candidates", func() {
		it("logs the candidate entries", func() {
			emitter.Candidates([]packit.BuildpackPlanEntry{
				{
					Version: "some-version",
					Metadata: map[string]interface{}{
						"version-source": "some-source",
					},
				},
				{Version: "other-version"},
			})
			Expect(buffer.String()).To(Equal(`    Candidate version sources (in priority order):
      some-source -> "some-version"
      <unknown>   -> "other-version"

`))
		})
	})

	context("SelectedDependency", func() {
		var (
			entry      packit.BuildpackPlanEntry
			dependency postal.Dependency
		)

		it.Before(func() {
			entry = packit.BuildpackPlanEntry{
				Metadata: map[string]interface{}{
					"version-source": "some-source",
				},
			}

			dependency = postal.Dependency{
				Name:    "Python",
				Version: "some-version",
			}

			var err error
			dependency.DeprecationDate, err = time.Parse(time.RFC3339, "2021-04-01T00:00:00Z")
			Expect(err).NotTo(HaveOccurred())
		})

		it("logs the selected dependency", func() {
			emitter.SelectedDependency(entry, dependency, time.Now())
			Expect(buffer.String()).To(Equal("    Selected Python version (using some-source): some-version\n\n"))
		})

		context("when it is within 30 days of the deprecation date", func() {
			var now time.Time

			it.Before(func() {
				now = dependency.DeprecationDate.Add(-29 * 24 * time.Hour)
			})

			it("returns a warning that the dependency will be deprecated after the deprecation date", func() {
				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected Python version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of Python will be deprecated after 2021-04-01.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of Python before this time.\n\n"))
			})
		})

		context("when it is on the the deprecation date", func() {
			var now time.Time

			it.Before(func() {
				now = dependency.DeprecationDate
			})

			it("returns a warning that the version of the dependency is no longer supported", func() {
				emitter.SelectedDependency(entry, dependency, now)
				Expect(buffer.String()).To(ContainSubstring("    Selected Python version (using some-source): some-version\n"))
				Expect(buffer.String()).To(ContainSubstring("      Version some-version of Python is deprecated.\n"))
				Expect(buffer.String()).To(ContainSubstring("      Migrate your application to a supported version of Python.\n\n"))
			})
		})
	})

	context("Environment", func() {
		it("prints details about the environment", func() {
			emitter.Environment(packit.Environment{
				"PYTHONPATH.override": "/some/path",
			})

			Expect(buffer.String()).To(ContainSubstring("  Configuring environment"))
			Expect(buffer.String()).To(ContainSubstring("    PYTHONPATH -> \"/some/path\""))
		})
	})
}
