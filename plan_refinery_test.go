package pythonruntime_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	pythonruntime "github.com/paketo-community/python-runtime"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPlanRefinery(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		planRefinery pythonruntime.PlanRefinery
	)

	it.Before(func() {
		planRefinery = pythonruntime.NewBuildPlanRefinery()
	})

	context("BillOfMaterial", func() {
		it("creates a buildpack plan entry from the given dependency", func() {
			entry := planRefinery.BillOfMaterials(postal.Dependency{
				ID:      "some-id",
				Name:    "some-name",
				Stacks:  []string{"some-stack"},
				URI:     "some-uri",
				SHA256:  "some-sha",
				Version: "some-version",
			})
			Expect(entry).To(Equal(packit.BuildpackPlanEntry{
				Name:    "some-id",
				Version: "some-version",
				Metadata: map[string]interface{}{
					"licenses": []string{},
					"name":     "some-name",
					"sha256":   "some-sha",
					"stacks":   []string{"some-stack"},
					"uri":      "some-uri",
				},
			},
			))
		})
	})
}
