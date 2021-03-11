package cpython

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

// BuildPlanRefinery generates a BuildpackPlan Entry containing the
// Bill-of-Materials of a given dependency.
type BuildPlanRefinery struct{}

// NewBuildPlanRefinery creates a BuildPlanRefinery.
func NewBuildPlanRefinery() BuildPlanRefinery {
	return BuildPlanRefinery{}
}

// BillOfMaterials generates a Bill-of-Materials describing buildpack's
// contributions to the app image.
func (r BuildPlanRefinery) BillOfMaterials(dependency postal.Dependency) packit.BuildpackPlanEntry {
	return packit.BuildpackPlanEntry{
		Name: dependency.ID,
		Metadata: map[string]interface{}{
			"licenses": []string{},
			"name":     dependency.Name,
			"sha256":   dependency.SHA256,
			"stacks":   dependency.Stacks,
			"uri":      dependency.URI,
			"version":  dependency.Version,
		},
	}
}
