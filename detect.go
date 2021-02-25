package cpython

import (
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface VersionParser --output fakes/version_parser.go
type VersionParser interface {
	ParseVersion(path string) (version string, err error)
}

type BuildPlanMetadata struct {
	Version       string `toml:"version"`
	VersionSource string `toml:"version-source"`
}

func Detect(buildpackYMLParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement
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
