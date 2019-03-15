package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/python-cnb/python"

	"github.com/cloudfoundry/libcfbuildpack/detect"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create a default detection context: %s", err)
		os.Exit(101)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	runtimePath := filepath.Join(context.Application.Root, "runtime.txt")
	exists, err := helper.FileExists(runtimePath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	var version string
	if exists {
		version, err = readRuntimeTxtVersion(runtimePath)
		if err != nil {
			return detect.FailStatusCode, err
		}
	}

	buildpackYAMLPath := filepath.Join(context.Application.Root, "buildpack.yml")
	exists, err = helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	if exists {
		version, err = readBuildpackYamlVersion(buildpackYAMLPath)
		if err != nil {
			return detect.FailStatusCode, err
		}
	}

	return context.Pass(buildplan.BuildPlan{
		python.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
	})
}

func readRuntimeTxtVersion(runtimePath string) (string, error) {
	buf, err := ioutil.ReadFile(runtimePath)
	return string(buf), err
}

func readBuildpackYamlVersion(buildpackYAMLPath string) (string, error) {
	buf, err := ioutil.ReadFile(buildpackYAMLPath)
	if err != nil {
		return "", err
	}

	config := struct {
		Python struct {
			Version string `yaml:"version"`
		} `yaml:"python"`
	}{}
	if err := yaml.Unmarshal(buf, &config); err != nil {
		return "", err
	}

	return config.Python.Version, nil
}
