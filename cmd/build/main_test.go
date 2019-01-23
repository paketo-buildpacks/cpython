package main

import (
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/google/go-cmp/cmp"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitBuild(t *testing.T) {
	spec.Run(t, "Build", testBuild, spec.Report(report.Terminal{}))
}

func testBuild(t *testing.T, _ spec.G, it spec.S) {
	it("always passes", func() {
		f := test.NewBuildFactory(t)
		code, err := runBuild(f.Build)
		if err != nil {
			t.Error("Err in build : ", err)
		}

		if diff := cmp.Diff(code, detect.PassStatusCode); diff != "" {
			t.Error("Problem : ", diff)
		}
	})
}
