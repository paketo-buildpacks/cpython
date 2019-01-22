package main

import (
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/google/go-cmp/cmp"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnit(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, _ spec.G, it spec.S) {
	it("always passes", func() {
		f := test.NewDetectFactory(t)
		code, err := runDetect(f.Detect)
		if err != nil {
			t.Error("Err in detect : ", err)
		}

		if diff := cmp.Diff(code, detect.PassStatusCode); diff != "" {
			t.Error("Problem : ", diff)
		}
	})
}
