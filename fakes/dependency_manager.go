package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/postal"
)

type DependencyManager struct {
	DeliverDependencyCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Dependency      cargo.ConfigMetadataDependency
			CnbPath         string
			DestinationPath string
			PlatformPath    string
		}
		Returns struct {
			Error error
		}
		Stub func(cargo.ConfigMetadataDependency, string, string, string) error
	}
	GenerateBillOfMaterialsCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Dependencies []postal.Dependency
		}
		Returns struct {
			BOMEntrySlice []packit.BOMEntry
		}
		Stub func(...postal.Dependency) []packit.BOMEntry
	}
}

func (f *DependencyManager) DeliverDependency(param1 cargo.ConfigMetadataDependency, param2 string, param3 string, param4 string) error {
	f.DeliverDependencyCall.mutex.Lock()
	defer f.DeliverDependencyCall.mutex.Unlock()
	f.DeliverDependencyCall.CallCount++
	f.DeliverDependencyCall.Receives.Dependency = param1
	f.DeliverDependencyCall.Receives.CnbPath = param2
	f.DeliverDependencyCall.Receives.DestinationPath = param3
	f.DeliverDependencyCall.Receives.PlatformPath = param4
	if f.DeliverDependencyCall.Stub != nil {
		return f.DeliverDependencyCall.Stub(param1, param2, param3, param4)
	}
	return f.DeliverDependencyCall.Returns.Error
}
func (f *DependencyManager) GenerateBillOfMaterials(param1 ...postal.Dependency) []packit.BOMEntry {
	f.GenerateBillOfMaterialsCall.mutex.Lock()
	defer f.GenerateBillOfMaterialsCall.mutex.Unlock()
	f.GenerateBillOfMaterialsCall.CallCount++
	f.GenerateBillOfMaterialsCall.Receives.Dependencies = param1
	if f.GenerateBillOfMaterialsCall.Stub != nil {
		return f.GenerateBillOfMaterialsCall.Stub(param1...)
	}
	return f.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice
}
