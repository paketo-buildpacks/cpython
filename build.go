package cpython

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

		logs.Process("Resolving CPython version")
		entry := entries.Resolve(context.Plan.Entries)
		entryVersion, _ := entry.Metadata["version"].(string)

		// This is done because the core dependencies pipeline provides the cpython
		// dependency under the name python.
		entry.Name = "python"

		dependency, err := dependencies.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, entryVersion, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		// This is done because the core dependencies pipeline provides the cpython
		// dependency under the name python.
		dependency.ID = "cpython"
		dependency.Name = "CPython"

		logs.SelectedDependency(entry, dependency, clock.Now())
		bom := planRefinery.BillOfMaterials(dependency)

		cpythonLayer, err := context.Layers.Get(Cpython)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedSHA, ok := cpythonLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logs.Process("Reusing cached layer %s", cpythonLayer.Path)
			logs.Break()

			return packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{bom},
				},
				Layers: []packit.Layer{cpythonLayer},
			}, nil
		}

		logs.Process("Executing build process")

		cpythonLayer, err = cpythonLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		cpythonLayer.Build = entry.Metadata["build"] == true
		cpythonLayer.Cache = entry.Metadata["build"] == true
		cpythonLayer.Launch = entry.Metadata["launch"] == true

		logs.Subprocess("Installing CPython %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencies.Install(dependency, context.CNBPath, cpythonLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logs.Action("Completed in %s", duration.Round(time.Millisecond))

		cpythonLayer.Metadata = map[string]interface{}{
			DepKey:     dependency.SHA256,
			"built_at": clock.Now().Format(time.RFC3339Nano),
		}

		os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Join(cpythonLayer.Path, "bin"), os.Getenv("PATH")))

		cpythonLayer.SharedEnv.Override("PYTHONPATH", cpythonLayer.Path)

		logs.Break()
		logs.Environment(cpythonLayer.SharedEnv)

		return packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{bom},
			},
			Layers: []packit.Layer{cpythonLayer},
		}, nil
	}
}
