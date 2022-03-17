package cpython

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface EntryResolver --output fakes/entry_resolver.go
//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go

// EntryResolver defines the interface for picking the most relevant entry from
// the Buildpack Plan entries.
type EntryResolver interface {
	Resolve(string, []packit.BuildpackPlanEntry, []interface{}) (packit.BuildpackPlanEntry, []packit.BuildpackPlanEntry)
	MergeLayerTypes(string, []packit.BuildpackPlanEntry) (launch, build bool)
}

// DependencyManager defines the interface for picking the best matching
// dependency installing it, and generating a BOM.
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, destinationPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right cpython dependency to install, install it in a
// layer, and generate Bill-of-Materials. It also makes use of the checksum of
// the dependency to reuse the layer when possible.
func Build(entries EntryResolver, dependencies DependencyManager, logs scribe.Emitter, clock chronos.Clock) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logs.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logs.Process("Resolving CPython version")

		entry, sortedEntries := entries.Resolve(Cpython, context.Plan.Entries, Priorities)
		logs.Candidates(sortedEntries)

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

		source, _ := entry.Metadata["version-source"].(string)
		if source == "buildpack.yml" {
			nextMajorVersion := semver.MustParse(context.BuildpackInfo.Version).IncMajor()
			logs.Subprocess("WARNING: Setting the CPython version through buildpack.yml is deprecated and will be removed in %s v%s.", context.BuildpackInfo.Name, nextMajorVersion.String())
			logs.Subprocess("Please specify the version through the $BP_CPYTHON_VERSION environment variable instead. See docs for more information.")
			logs.Break()
		}

		bom := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := entries.MergeLayerTypes(Cpython, context.Plan.Entries)

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = bom
		}

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = bom
		}

		cpythonLayer, err := context.Layers.Get(Cpython)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cpythonLayer.Launch, cpythonLayer.Build, cpythonLayer.Cache = launch, build, build

		cachedSHA, ok := cpythonLayer.Metadata[DepKey].(string)
		if ok && cachedSHA == dependency.SHA256 {
			logs.Process("Reusing cached layer %s", cpythonLayer.Path)
			logs.Break()

			return packit.BuildResult{
				Layers: []packit.Layer{cpythonLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		logs.Process("Executing build process")

		cpythonLayer, err = cpythonLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		cpythonLayer.Launch, cpythonLayer.Build, cpythonLayer.Cache = launch, build, build

		logs.Subprocess("Installing CPython %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencies.Deliver(dependency, context.CNBPath, cpythonLayer.Path, context.Platform.Path)
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
		logs.Process("Configuring environment")
		logs.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(cpythonLayer.SharedEnv))
		logs.Break()

		return packit.BuildResult{
			Layers: []packit.Layer{cpythonLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
