package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	cpython "github.com/paketo-community/cpython"
)

func main() {
	buildpackYMLParser := cpython.NewBuildpackYMLParser()

	dependencies := postal.NewService(cargo.NewTransport())
	planRefinery := cpython.NewBuildPlanRefinery()
	logs := cpython.NewLogEmitter(os.Stdout)
	entries := cpython.NewPlanEntryResolver(logs)

	packit.Run(
		cpython.Detect(buildpackYMLParser),
		cpython.Build(entries, dependencies, planRefinery, logs, chronos.DefaultClock),
	)
}
