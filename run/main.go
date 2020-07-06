package main

import (
	"os"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	pythonruntime "github.com/paketo-community/python-runtime"
)

func main() {
	buildpackYMLParser := pythonruntime.NewBuildpackYMLParser()

	dependencies := postal.NewService(cargo.NewTransport())
	planRefinery := pythonruntime.NewBuildPlanRefinery()
	logs := pythonruntime.NewLogEmitter(os.Stdout)
	entries := pythonruntime.NewPlanEntryResolver(logs)

	packit.Run(
		pythonruntime.Detect(buildpackYMLParser),
		pythonruntime.Build(entries, dependencies, planRefinery, logs, chronos.DefaultClock),
	)
}
