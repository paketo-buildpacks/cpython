package cpython

import (
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface PythonInstaller --output fakes/installer.go
//go:generate faux --interface PythonPipCleanup --output fakes/pip_cleanup.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

// DependencyManager defines the interface for picking the best matching
// dependency installing it, and generating a BOM.
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, destinationPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

// PythonInstaller defines the interface for installing python from source
type PythonInstaller interface {
	Install(
		sourcePath string,
		workingDir string,
		entry packit.BuildpackPlanEntry,
		dependency postal.Dependency,
		layerPath string,
	) error
}

// PythonPipCleanup defines the interface for cleaning up pip after a python installation
type PythonPipCleanup interface {
	Cleanup(packages []string, targetLayer string) error
}

type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right cpython dependency to install, install it in a
// layer, and generate Bill-of-Materials. It also makes use of the checksum of
// the dependency to reuse the layer when possible.
func Build(
	dependencies DependencyManager,
	pythonInstaller PythonInstaller,
	pipCleanup PythonPipCleanup,
	sbomGenerator SBOMGenerator,
	logger scribe.Emitter,
	clock chronos.Clock,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logger.Process("Resolving CPython version")

		planner := draft.NewPlanner()

		entry, sortedEntries := planner.Resolve(Cpython, context.Plan.Entries, Priorities)
		logger.Candidates(sortedEntries)

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

		logger.SelectedDependency(entry, dependency, clock.Now())

		legacySBOM := dependencies.GenerateBillOfMaterials(dependency)
		launch, build := planner.MergeLayerTypes(Cpython, context.Plan.Entries)

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = legacySBOM
		}

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = legacySBOM
		}

		cpythonLayer, err := context.Layers.Get(Cpython)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cpythonLayer.Launch, cpythonLayer.Build, cpythonLayer.Cache = launch, build, build

		cachedChecksum, ok := cpythonLayer.Metadata[DepKey].(string)
		dependencyChecksum := dependency.Checksum
		if dependencyChecksum == "" {
			//nolint:staticcheck // SHA256 is only a fallback in case Checksum is not present
			dependencyChecksum = dependency.SHA256
		}

		if ok && cachedChecksum != "" && cargo.Checksum(cachedChecksum).MatchString(dependencyChecksum) {
			logger.Process("Reusing cached layer %s", cpythonLayer.Path)
			logger.Break()

			return packit.BuildResult{
				Layers: []packit.Layer{cpythonLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		logger.Process("Executing build process")

		cpythonLayer, err = cpythonLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		cpythonLayer.Launch, cpythonLayer.Build, cpythonLayer.Cache = launch, build, build

		logger.Subprocess("Installing CPython %s", dependency.Version)

		var duration time.Duration

		// Install python from source when URI and Source match
		if dependency.URI == dependency.Source {
			sourcePath := filepath.Join(cpythonLayer.Path, "python-source")

			if dependency.StripComponents == 0 {
				// CPython distributions have one layer of directory prefix, so strip that when unpacking
				dependency.StripComponents = 1
			}

			err = os.Mkdir(sourcePath, 0755)
			if err != nil {
				// untested - hard to force a test failure
				return packit.BuildResult{}, err
			}

			downloadDuration, err := clock.Measure(func() error {
				return dependencies.Deliver(dependency, context.CNBPath, sourcePath, context.Platform.Path)
			})
			if err != nil {
				return packit.BuildResult{}, err
			}

			installDuration, err := clock.Measure(func() error {
				return pythonInstaller.Install(sourcePath, context.WorkingDir, entry, dependency, cpythonLayer.Path)
			})
			if err != nil {
				return packit.BuildResult{}, err
			}

			duration = downloadDuration + installDuration

			err = os.RemoveAll(sourcePath)
			if err != nil {
				return packit.BuildResult{}, err
			}
		} else {
			// Otherwise extract context from URI into layer
			downloadDuration, err := clock.Measure(func() error {
				return dependencies.Deliver(dependency, context.CNBPath, cpythonLayer.Path, context.Platform.Path)
			})
			if err != nil {
				return packit.BuildResult{}, err
			}

			duration = downloadDuration
		}

		pipCleanupDuration, err := clock.Measure(func() error {
			if _, ok := os.LookupEnv("BP_CPYTHON_RM_SETUPTOOLS"); ok {
				return pipCleanup.Cleanup(pipPackagesToBeUninstalled(), cpythonLayer.Path)
			}
			return nil
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		duration += pipCleanupDuration

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.GeneratingSBOM(cpythonLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, cpythonLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		cpythonLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cpythonLayer.Metadata = map[string]interface{}{
			DepKey: dependencyChecksum,
		}

		cpythonLayer.BuildEnv.Default("PYTHONPYCACHEPREFIX", "/tmp")
		cpythonLayer.SharedEnv.Default("PYTHONPATH", cpythonLayer.Path)
		cpythonLayer.ExecD = []string{filepath.Join(context.CNBPath, "bin", "env")}

		if exists, err := fs.Exists(filepath.Join(cpythonLayer.Path, "bin", "python")); err != nil {
			return packit.BuildResult{}, err
		} else if exists {
			logger.Debug.Detail("bin/python already exists")
		} else {
			logger.Debug.Detail("Writing symlink bin/python")
			if err := os.Symlink(filepath.Join(cpythonLayer.Path, "bin", "python3"), filepath.Join(cpythonLayer.Path, "bin", "python")); err != nil {
				return packit.BuildResult{}, err
			}
		}

		logger.Break()

		logger.EnvironmentVariables(cpythonLayer)

		return packit.BuildResult{
			Layers: []packit.Layer{cpythonLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
