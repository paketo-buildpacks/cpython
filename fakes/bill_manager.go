package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/postal"
)

type BillManager struct {
	BillOfMaterialsCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependency postal.Dependency
		}
		Returns struct {
			BOMEntry packit.BOMEntry
		}
		Stub func(postal.Dependency) packit.BOMEntry
	}
}

func (f *BillManager) BillOfMaterials(param1 postal.Dependency) packit.BOMEntry {
	f.BillOfMaterialsCall.Lock()
	defer f.BillOfMaterialsCall.Unlock()
	f.BillOfMaterialsCall.CallCount++
	f.BillOfMaterialsCall.Receives.Dependency = param1
	if f.BillOfMaterialsCall.Stub != nil {
		return f.BillOfMaterialsCall.Stub(param1)
	}
	return f.BillOfMaterialsCall.Returns.BOMEntry
}
