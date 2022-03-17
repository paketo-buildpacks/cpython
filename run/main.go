package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	cpython "github.com/paketo-buildpacks/cpython"
)

func main() {
	entries := draft.NewPlanner()
	dependencies := postal.NewService(cargo.NewTransport())
	buildpackYMLParser := cpython.NewBuildpackYMLParser()
	logs := scribe.NewEmitter(os.Stdout)

	packit.Run(
		cpython.Detect(buildpackYMLParser),
		cpython.Build(entries, dependencies, logs, chronos.DefaultClock),
	)
}
