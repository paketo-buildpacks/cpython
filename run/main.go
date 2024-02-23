package main

import (
	"fmt"
	"os"

	cpython "github.com/paketo-buildpacks/cpython"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type Generator struct{}

func (f Generator) GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error) {
	return sbom.GenerateFromDependency(dependency, path)
}

func main() {
	fmt.Println("***************************************** Hello World")
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	dependencies := postal.NewService(cargo.NewTransport())
	pythonSourceInstaller := cpython.NewCPythonInstaller(
		pexec.NewExecutable("configure"),
		pexec.NewExecutable("make"),
		logger,
	)
	sbomGenerator := Generator{}

	packit.Run(
		cpython.Detect(),
		cpython.Build(
			dependencies,
			pythonSourceInstaller,
			sbomGenerator,
			logger,
			chronos.DefaultClock),
	)
}
