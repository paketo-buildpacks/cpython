package cpython

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

// BuildPlanMetadata is the buildpack specific data included in build plan
// requirements.
type BuildPlanMetadata struct {
	// Version denotes the version constraint to be requested in the requirements.
	Version string `toml:"version"`

	// VersionSource denotes the source of the version information. This may be
	// used by the consumer of the metadata to determine the priority of this
	// version request.
	VersionSource string `toml:"version-source"`
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detect always passes, and will contribute a Build Plan that provides cpython.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		// This is to comply with RFC 1:
		// https://github.com/paketo-community/cpython/blob/main/rfcs/0001-buildpack-yml-to-env-var.md
		_, err := os.Stat(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err == nil {
			errMessage := "buildpack.yml detected. Configuration via buildpack.yml is unsupported from v1.0.0. Please remove the buildpack.yml and use $BP_CPYTHON_VERSION environment variable to specify the version. See https://github.com/paketo-community/cpython/blob/main/README.md for more information."
			return packit.DetectResult{}, packit.Fail.WithMessage(errMessage)
		} else if !errors.Is(err, os.ErrNotExist) {
			return packit.DetectResult{}, err
		}

		if version, ok := os.LookupEnv("BP_CPYTHON_VERSION"); ok {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Cpython,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: "BP_CPYTHON_VERSION",
				},
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
