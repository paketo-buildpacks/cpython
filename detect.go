package cpython

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go

// VersionParser defines the interface for determining the version of Cpython.
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

// BuildPlanMetadata is the buildpack specific data included in build plan
// requirements.
type BuildPlanMetadata struct {
	// Version denotes the version constraint to be requested in the requirements.
	Version string `toml:"version"`

	// VersionSource denotes the source of the version information. This may be
	// used by the consumer of the metadata to determine the priority of this
	// version request.
	VersionSource string `toml:"version-source"`

	// ConfigureFlags denotes the configure flags to be requested in the requirements.
	// This is used to run configure before make and make install.
	ConfigureFlags string `toml:"configure-flags"`
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detect always passes, and will contribute a Build Plan that provides cpython.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		bpYML, err := fs.Exists(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("failed to check for buildpack.yml: %w", err)
		}
		if bpYML {
			return packit.DetectResult{}, fmt.Errorf("working directory contains deprecated 'buildpack.yml'; use environment variables for configuration")
		}

		var requirements []packit.BuildPlanRequirement

		if version, ok := os.LookupEnv("BP_CPYTHON_VERSION"); ok {
			metadata := BuildPlanMetadata{
				Version:       version,
				VersionSource: "BP_CPYTHON_VERSION",
			}

			// Ignored for stacks that do not build python from source
			if flags, ok := os.LookupEnv("BP_CPYTHON_CONFIGURE_FLAGS"); ok {
				metadata.ConfigureFlags = flags
			}

			requirements = append(requirements, packit.BuildPlanRequirement{
				Name:     Cpython,
				Metadata: metadata,
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Cpython},
				},
				Requires: requirements,
			},
		}, nil
	}
}
