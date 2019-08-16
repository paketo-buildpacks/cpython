package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/python-cnb/python"

	"github.com/cloudfoundry/libcfbuildpack/detect"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

type Config struct {
	Version string `yaml:"version"`
}

type BuildpackYAML struct {
	Config Config `yaml:"python"`
}

func runDetect(context detect.Detect) (int, error) {
	var version string

	buildpackYAMLPath := filepath.Join(context.Application.Root, "buildpack.yml")
	buildpackYAMLExists, err := helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	if buildpackYAMLExists {
		buildpackYAML := BuildpackYAML{}
		err = helper.ReadBuildpackYaml(buildpackYAMLPath, &buildpackYAML)
		if err != nil {
			return detect.FailStatusCode, err
		}

		if buildpackYAML.Config.Version != "" {
			version = buildpackYAML.Config.Version
		}
	}

	if version != "" {
		return context.Pass(buildplan.Plan{
			Provides: []buildplan.Provided{{Name: python.Dependency}},
			Requires: []buildplan.Required{{
				Name:     python.Dependency,
				Version:  version,
				Metadata: buildplan.Metadata{"launch": true},
			}},
		})
	}

	return context.Pass(buildplan.Plan{
		Provides: []buildplan.Provided{{Name: python.Dependency}},
	})
}
