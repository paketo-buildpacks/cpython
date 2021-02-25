package cpython_test

import (
	"bytes"
	"testing"

	"github.com/paketo-buildpacks/packit"
	cpython "github.com/paketo-community/cpython"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLogEmitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buffer  *bytes.Buffer
		emitter cpython.LogEmitter
	)

	it.Before(func() {
		buffer = bytes.NewBuffer(nil)
		emitter = cpython.NewLogEmitter(buffer)
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
					Metadata: map[string]interface{}{
						"version":        "some-version",
						"version-source": "some-source",
					},
				},
				{
					Metadata: map[string]interface{}{
						"version": "other-version",
					},
				},
			})
			Expect(buffer.String()).To(Equal(`    Candidate version sources (in priority order):
      some-source -> "some-version"
      <unknown>   -> "other-version"

`))
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
