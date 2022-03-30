package main

import (
	"os"

	cpython "github.com/paketo-buildpacks/cpython"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	dependencies := postal.NewService(cargo.NewTransport())
	buildpackYMLParser := cpython.NewBuildpackYMLParser()
	logs := scribe.NewEmitter(os.Stdout)

	packit.Run(
		cpython.Detect(buildpackYMLParser),
		cpython.Build(dependencies, logs, chronos.DefaultClock),
	)
}
