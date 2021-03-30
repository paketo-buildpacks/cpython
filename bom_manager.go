package cpython

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

// BOMManager generates a Bill-of-Materials entry.
type BOMManager struct{}

// NewBOMManager creates a BOMManager.
func NewBOMManager() BOMManager {
	return BOMManager{}
}

// BillOfMaterials generates a Bill-of-Materials entry describing buildpack's
// contributions to the app image given the dependency it installs.
func (r BOMManager) BillOfMaterials(dependency postal.Dependency) packit.BOMEntry {
	return packit.BOMEntry{
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
