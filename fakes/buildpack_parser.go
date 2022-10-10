package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type BuildpackParser struct {
	ParseCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Path string
		}
		Returns struct {
			Config cargo.Config
			Error  error
		}
		Stub func(string) (cargo.Config, error)
	}
}

func (f *BuildpackParser) Parse(param1 string) (cargo.Config, error) {
	f.ParseCall.mutex.Lock()
	defer f.ParseCall.mutex.Unlock()
	f.ParseCall.CallCount++
	f.ParseCall.Receives.Path = param1
	if f.ParseCall.Stub != nil {
		return f.ParseCall.Stub(param1)
	}
	return f.ParseCall.Returns.Config, f.ParseCall.Returns.Error
}
