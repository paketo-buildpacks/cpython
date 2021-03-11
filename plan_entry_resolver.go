package cpython

import (
	"sort"

	"github.com/paketo-buildpacks/packit"
)

// PlanEntryResolver picks the most relevant entry from the Buildpack Plan
// entries.
type PlanEntryResolver struct {
	logger LogEmitter
}

// NewPlanEntryResolver creates a PlanEntryResolver.
func NewPlanEntryResolver(logger LogEmitter) PlanEntryResolver {
	return PlanEntryResolver{
		logger: logger,
	}
}

// Resolve goes through all the given Buildpack Plan entries, priority-sorts
// them and picks the highest priority entry. It also merges the metadata from
// all entries into the chosen entry.
func (r PlanEntryResolver) Resolve(entries []packit.BuildpackPlanEntry) packit.BuildpackPlanEntry {
	var (
		priorities = map[string]int{
			"buildpack.yml": 3,
			"":              -1,
		}
	)

	sort.Slice(entries, func(i, j int) bool {
		leftSource := entries[i].Metadata["version-source"]
		left, _ := leftSource.(string)

		rightSource := entries[j].Metadata["version-source"]
		right, _ := rightSource.(string)

		return priorities[left] > priorities[right]
	})

	chosenEntry := entries[0]

	if chosenEntry.Metadata == nil {
		chosenEntry.Metadata = map[string]interface{}{}
	}

	for _, entry := range entries {
		if entry.Metadata["build"] == true {
			chosenEntry.Metadata["build"] = true
		}
		if entry.Metadata["launch"] == true {
			chosenEntry.Metadata["launch"] = true
		}
	}

	r.logger.Candidates(entries)

	return chosenEntry
}
