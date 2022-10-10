package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/v2"
)

type PythonInstaller struct {
	InstallCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			SourcePath        string
			WorkingDir        string
			Entry             packit.BuildpackPlanEntry
			DependencyVersion string
			LayerPath         string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string, packit.BuildpackPlanEntry, string, string) error
	}
}

func (f *PythonInstaller) Install(param1 string, param2 string, param3 packit.BuildpackPlanEntry, param4 string, param5 string) error {
	f.InstallCall.mutex.Lock()
	defer f.InstallCall.mutex.Unlock()
	f.InstallCall.CallCount++
	f.InstallCall.Receives.SourcePath = param1
	f.InstallCall.Receives.WorkingDir = param2
	f.InstallCall.Receives.Entry = param3
	f.InstallCall.Receives.DependencyVersion = param4
	f.InstallCall.Receives.LayerPath = param5
	if f.InstallCall.Stub != nil {
		return f.InstallCall.Stub(param1, param2, param3, param4, param5)
	}
	return f.InstallCall.Returns.Error
}
