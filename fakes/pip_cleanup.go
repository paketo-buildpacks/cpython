package fakes

import "sync"

type PythonPipCleanup struct {
	CleanupCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Packages    []string
			TargetLayer string
		}
		Returns struct {
			Error error
		}
		Stub func([]string, string) error
	}
}

func (f *PythonPipCleanup) Cleanup(param1 []string, param2 string) error {
	f.CleanupCall.mutex.Lock()
	defer f.CleanupCall.mutex.Unlock()
	f.CleanupCall.CallCount++
	f.CleanupCall.Receives.Packages = param1
	f.CleanupCall.Receives.TargetLayer = param2
	if f.CleanupCall.Stub != nil {
		return f.CleanupCall.Stub(param1, param2)
	}
	return f.CleanupCall.Returns.Error
}
