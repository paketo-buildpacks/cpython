package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/postal"
)

type PythonInstaller struct {
	InstallCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			SourcePath string
			WorkingDir string
			Entry      packit.BuildpackPlanEntry
			Dependency postal.Dependency
			LayerPath  string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, packit.BuildpackPlanEntry, postal.Dependency, string) error
	}
}

func (f *PythonInstaller) Install(param1 string, param2 string, param3 packit.BuildpackPlanEntry, param4 postal.Dependency, param5 string) error {
	f.InstallCall.mutex.Lock()
	defer f.InstallCall.mutex.Unlock()
	f.InstallCall.CallCount++
	f.InstallCall.Receives.SourcePath = param1
	f.InstallCall.Receives.WorkingDir = param2
	f.InstallCall.Receives.Entry = param3
	f.InstallCall.Receives.Dependency = param4
	f.InstallCall.Receives.LayerPath = param5
	if f.InstallCall.Stub != nil {
		return f.InstallCall.Stub(param1, param2, param3, param4, param5)
	}
	return f.InstallCall.Returns.Error
}
