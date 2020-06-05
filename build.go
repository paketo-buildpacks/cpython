package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
type EntryResolver interface {
	Resolve([]packit.BuildpackPlanEntry) packit.BuildpackPlanEntry
}

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Install(dependency postal.Dependency, cnbPath, layerPath string) error
}

//go:generate faux --interface PlanRefinery --output fakes/plan_refinery.go
type PlanRefinery interface {
	BillOfMaterials(dependency postal.Dependency) packit.BuildpackPlanEntry
}

func Build(entries EntryResolver, dependencies DependencyManager, planRefinery PlanRefinery, logs LogEmitter, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logs.Title(context.BuildpackInfo)

		logs.Process("Resolving Python version")
		entry := entries.Resolve(context.Plan.Entries)

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, entry.Version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logs.SelectedDependency(entry, dependency, clock.Now())
		bom := planRefinery.BillOfMaterials(dependency)

		pythonLayer, err := context.Layers.Get(Python)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedSHA, ok := pythonLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logs.Process("Reusing cached layer %s", pythonLayer.Path)
			logs.Break()

			return packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{bom},
				},
				Layers: []packit.Layer{pythonLayer},
			}, nil
		}

		logs.Process("Executing build process")

		err = pythonLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		pythonLayer.Build = entry.Metadata["build"] == true
		pythonLayer.Cache = entry.Metadata["build"] == true
		pythonLayer.Launch = entry.Metadata["launch"] == true

		logs.Subprocess("Installing Python %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencies.Install(dependency, context.CNBPath, pythonLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logs.Action("Completed in %s", duration.Round(time.Millisecond))

		pythonLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Join(pythonLayer.Path, "bin"), os.Getenv("PATH")))

		pythonLayer.SharedEnv.Override("PYTHONPATH", pythonLayer.Path)

		logs.Break()
		logs.Environment(pythonLayer.SharedEnv)

		return packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{bom},
			},
			Layers: []packit.Layer{pythonLayer},
		}, nil
	}
}
