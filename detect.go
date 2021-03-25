package cpython

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
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
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detect always passes, and will contribute a Build Plan that provides cpython.
func Detect(buildpackYMLParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		// TODO(restructure): Remove legacy requirements
		var requirementsLegacy []packit.BuildPlanRequirement

		version, err := buildpackYMLParser.ParseVersion(filepath.Join(context.WorkingDir, "buildpack.yml"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if version != "" {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Cpython,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: "buildpack.yml",
				},
			})
			requirementsLegacy = append(requirementsLegacy, packit.BuildPlanRequirement{
				Name: Python,
				Metadata: BuildPlanMetadata{
					Version:       version,
					VersionSource: "buildpack.yml",
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Cpython},
				},
				Requires: requirements,
				Or: []packit.BuildPlan{
					{
						Provides: []packit.BuildPlanProvision{
							{Name: Python},
						},
						Requires: requirementsLegacy,
					},
				},
			},
		}, nil
	}
}
