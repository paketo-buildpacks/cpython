package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/draft"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
	cpython "github.com/paketo-community/cpython"
)

func main() {
	entries := draft.NewPlanner()
	dependencies := postal.NewService(cargo.NewTransport())
	logs := scribe.NewEmitter(os.Stdout)

	packit.Run(
		cpython.Detect(),
		cpython.Build(entries, dependencies, logs, chronos.DefaultClock),
	)
}
