package cpython_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
	cpython "github.com/paketo-community/cpython"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBOMManager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		bomManager cpython.BillManager
	)

	it.Before(func() {
		bomManager = cpython.NewBOMManager()
	})

	context("BillOfMaterial", func() {
		it("creates a buildpack plan entry from the given dependency", func() {
			entry := bomManager.BillOfMaterials(postal.Dependency{
				ID:      "some-id",
				Name:    "some-name",
				Stacks:  []string{"some-stack"},
				URI:     "some-uri",
				SHA256:  "some-sha",
				Version: "some-version",
			})

			Expect(entry).To(Equal(packit.BOMEntry{
				Name: "some-id",
				Metadata: map[string]interface{}{
					"licenses": []string{},
					"name":     "some-name",
					"sha256":   "some-sha",
					"stacks":   []string{"some-stack"},
					"uri":      "some-uri",
					"version":  "some-version",
				},
			},
			))
		})
	})
}
