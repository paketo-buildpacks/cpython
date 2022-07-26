package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"golang.org/x/exp/slices"
)

func ParseBuildpackToml(buildpackTomlPath string) cargo.Config {
	configParser := cargo.NewBuildpackParser()
	config, err := configParser.Parse(buildpackTomlPath)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s: %s", buildpackTomlPath, err))
	}
	return config
}

type RetrievalOutput struct {
	Versions []string
	ID       string
	Name     string
}

func GetNewVersions(id, name, stack string, buildpackTomlPath string, allVersions []*semver.Version) []string {
	config := ParseBuildpackToml(buildpackTomlPath)
	buildpackVersions := GetBuildpackVersions(id, stack, config)
	versionsFilteredByConstraints := FilterToConstraints(id, config, allVersions)
	versionsFilteredByPatches := FilterToPatches(versionsFilteredByConstraints, config, buildpackVersions)

	if len(versionsFilteredByPatches) < 1 {
		panic("No versions found")
	}

	return versionsFilteredByPatches
}

func GetBuildpackVersions(id, stack string, config cargo.Config) []string {
	var buildpackVersions []string
	for _, d := range config.Metadata.Dependencies {
		if d.ID != id || !slices.Contains(d.Stacks, stack) {
			continue
		}
		buildpackVersions = append(buildpackVersions, d.Version)
	}
	return buildpackVersions
}

func FilterToPatches(versionsFilteredByConstraints map[string][]*semver.Version, config cargo.Config, buildpackVersions []string) []string {
	var versionsToAdd []*semver.Version
	for constraintId, versions := range versionsFilteredByConstraints {
		var buildpackConstraint cargo.ConfigMetadataDependencyConstraint
		for _, constraint := range config.Metadata.DependencyConstraints {
			if constraint.ID == constraintId {
				buildpackConstraint = constraint
			}
		}

		sort.Slice(versions, func(i, j int) bool {
			return versions[i].LessThan(versions[j])
		})

		// if there are more requested patches than matching dependencies, just
		// return all matching dependencies.
		if buildpackConstraint.Patches > len(versions) {
			continue
		}

		// Buildpack.toml dependencies are usually in order from lowest to highest
		// version. We want to return the the n largest matching dependencies in the
		// same order, n being the constraint.Patches field from the buildpack.toml.
		// Here, we are returning the n highest matching Dependencies.
		versionsToAdd = append(versionsToAdd, versions[len(versions)-buildpackConstraint.Patches:]...)
	}

	var versionsAsStrings []string
	for _, version := range versionsToAdd {
		versionAsString := version.String()
		if !slices.Contains(buildpackVersions, versionAsString) {
			versionsAsStrings = append(versionsAsStrings, versionAsString)
		}
	}

	return versionsAsStrings
}

func FilterToConstraints(id string, config cargo.Config, allVersions []*semver.Version) map[string][]*semver.Version {
	semverConstraints := make(map[string]*semver.Constraints)
	for _, constraint := range config.Metadata.DependencyConstraints {
		if constraint.ID != id {
			continue
		}

		semverConstraint, err := semver.NewConstraint(constraint.Constraint)
		if err != nil {
			panic(err)
		}
		semverConstraints[constraint.ID] = semverConstraint
	}

	newVersions := make(map[string][]*semver.Version)
	for _, version := range allVersions {
		for constraintId, constraint := range semverConstraints {
			if constraint.Check(version) {
				newVersions[constraintId] = append(newVersions[constraintId], version)
			}
		}
	}
	return newVersions
}

type githubReleaseNamesDTO struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
}

func GetReleasesFromGithub(githubToken, org, repo string) []*semver.Version {
	client := &http.Client{}

	perPage := 100
	page := 1

	var allVersions []*semver.Version

	for ; ; page++ {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=%d&page=%d", org, repo, perPage, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if githubToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
		}

		res, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}

		err = res.Body.Close()
		if err != nil {
			panic(err)
		}

		var githubReleaseNames []githubReleaseNamesDTO
		err = json.Unmarshal(body, &githubReleaseNames)
		if err != nil {
			panic(err)
		}

		for _, release := range githubReleaseNames {
			allVersions = append(allVersions, sanitizeGithubReleaseName(release))
		}

		if len(githubReleaseNames) < perPage {
			break
		}
	}

	sort.Slice(allVersions, func(i, j int) bool {
		return allVersions[i].GreaterThan(allVersions[j])
	})

	return allVersions
}

func sanitizeGithubReleaseName(release githubReleaseNamesDTO) *semver.Version {
	name := strings.TrimSpace(release.Name)

	if strings.HasPrefix(name, "v") {
		name = strings.TrimPrefix(name, "v")
	}

	version, err := semver.NewVersion(name)
	if err == nil {
		return version
	}

	name = strings.TrimSpace(release.TagName)
	if strings.HasPrefix(name, "v") {
		name = strings.TrimPrefix(name, "v")
	}

	version, _ = semver.NewVersion(name)
	return version
}
