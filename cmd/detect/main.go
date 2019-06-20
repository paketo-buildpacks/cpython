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

	if err := context.BuildPlan.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Build Plan: %s\n", err)
		os.Exit(101)
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
	buildpackYAMLPath := filepath.Join(context.Application.Root, "buildpack.yml")
	exists, err := helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return detect.FailStatusCode, err
	}
	buildpackYAML := BuildpackYAML{}

	version := context.BuildPlan[python.Dependency].Version
	if exists {
		err = helper.ReadBuildpackYaml(buildpackYAMLPath, &buildpackYAML)
		if err != nil {
			return detect.FailStatusCode, err
		}
		if buildpackYAML.Config.Version != "" {
			version = buildpackYAML.Config.Version
		}
	}

	return context.Pass(buildplan.BuildPlan{
		python.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
	})
}
