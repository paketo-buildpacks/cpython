package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYMLParser := NewBuildpackYMLParser()

	dependencies := postal.NewService(cargo.NewTransport())
	planRefinery := NewBuildPlanRefinery()
	logs := NewLogEmitter(os.Stdout)
	entries := NewPlanEntryResolver(logs)

	packit.Run(
		Detect(buildpackYMLParser),
		Build(entries, dependencies, planRefinery, logs, chronos.DefaultClock),
	)

}
