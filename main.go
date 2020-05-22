package main

import (
	"os"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/postal"
)

func main() {
	buildpackYMLParser := NewBuildpackYMLParser()

	entries := NewPlanEntryResolver()
	dependencies := postal.NewService(cargo.NewTransport())
	planRefinery := NewBuildPlanRefinery()
	clock := NewClock(time.Now)
	logs := NewLogEmitter(os.Stdout)

	packit.Run(
		Detect(buildpackYMLParser),
		Build(entries, dependencies, planRefinery, clock, logs),
	)

}
