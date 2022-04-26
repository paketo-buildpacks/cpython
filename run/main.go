package main

import (
	"os"

	cpython "github.com/paketo-buildpacks/cpython"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type Generator struct{}

func (f Generator) GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error) {
	return sbom.GenerateFromDependency(dependency, path)
}

func main() {
	dependencies := postal.NewService(cargo.NewTransport())
	buildpackYMLParser := cpython.NewBuildpackYMLParser()
	sbomGenerator := Generator{}
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	packit.Run(
		cpython.Detect(buildpackYMLParser),
		cpython.Build(dependencies, sbomGenerator, logger, chronos.DefaultClock),
	)
}
