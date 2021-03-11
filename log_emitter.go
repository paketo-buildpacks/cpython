package cpython

import (
	"io"
	"strconv"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

// LogEmitter performs the logging during the build process.
type LogEmitter struct {
	scribe.Emitter
}

// NewLogEmitter creates a LogEmitter with output set to the given io.Writer.
func NewLogEmitter(output io.Writer) LogEmitter {
	return LogEmitter{
		Emitter: scribe.NewEmitter(output),
	}
}

// Title logs the name and version of the buildpack.
func (l LogEmitter) Title(info packit.BuildpackInfo) {
	l.Logger.Title("%s %s", info.Name, info.Version)
}

// Candidate logs the versions and version sources of all BuildpackPlan
// entries.
func (l LogEmitter) Candidates(entries []packit.BuildpackPlanEntry) {
	l.Subprocess("Candidate version sources (in priority order):")

	var (
		sources [][2]string
		maxLen  int
	)

	for _, entry := range entries {
		versionSource, ok := entry.Metadata["version-source"].(string)
		if !ok {
			versionSource = "<unknown>"
		}

		version, ok := entry.Metadata["version"].(string)
		if !ok {
			version = ""
		}

		if len(versionSource) > maxLen {
			maxLen = len(versionSource)
		}

		sources = append(sources, [2]string{versionSource, version})
	}

	for _, source := range sources {
		l.Action(("%-" + strconv.Itoa(maxLen) + "s -> %q"), source[0], source[1])
	}

	l.Break()
}

// Environment logs environment variables set by the buildpack.
func (l LogEmitter) Environment(env packit.Environment) {
	l.Process("Configuring environment")
	l.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(env))
	l.Break()
}
