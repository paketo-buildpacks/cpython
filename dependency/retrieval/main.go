package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/collections"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"golang.org/x/exp/slices"
)

type StackAndTargetPair struct {
	stacks []string
	target string
}

var supportedStacks = []StackAndTargetPair{
	{stacks: []string{"io.buildpacks.stacks.noble"}, target: "noble"},
	{stacks: []string{"io.buildpacks.stacks.jammy"}, target: "jammy"},
	{stacks: []string{"io.buildpacks.stacks.bionic"}, target: "bionic"},
	{stacks: []string{"*"}, target: "NONE"},
}

type PlatformStackTarget struct {
	stacks []string
	target string
	os string
	arch string
}

var supportedPlatforms = map[string][]string{
	"linux": {"amd64", "arm64"},
}

func getSuportedPlatformStackTargets() []PlatformStackTarget {
	var platformStackTargets []PlatformStackTarget

	for os, architectures := range supportedPlatforms {
		for _, arch := range architectures {
			for _, pair := range supportedStacks {
				platformStackTargets = append(platformStackTargets, PlatformStackTarget{
					stacks: pair.stacks,
					target: pair.target,
					os:     os,
					arch:   arch,
				})
			}
		}
	}

	return platformStackTargets
}

func getAsString(url string) (string, error) {
	response, err := http.DefaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not get project metadata: %w", err)
	}

	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response: %w", err)
	}

	return string(body), nil
}

func getAllVersions() (versionology.VersionFetcherArray, error) {
	body, err := getAsString("https://www.python.org/downloads/")
	if err != nil {
		return nil, fmt.Errorf("could not retrieve new versions from upstream: %w", err)
	}

	re := regexp.MustCompile(`release-number.*Python ([\d]+\.[\d]+\.[\d]+)`)

	var versions []string
	for _, line := range strings.Split(body, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			versions = append(versions, matches[1])
		}
	}

	return versionology.NewSimpleVersionFetcherArray(versions...)
}

func generateMetadata(versionFetcher versionology.VersionFetcher) ([]versionology.Dependency, error) {
	version := versionFetcher.Version().String()
	sourceURL := fmt.Sprintf("https://www.python.org/ftp/python/%[1]s/Python-%[1]s.tgz", version)

	sourceSHA256, err := getSha256(sourceURL, version)
	if err != nil {
		return nil, err
	}

	eolDate, err := getEOL(versionFetcher.Version())
	if err != nil {
		return nil, err
	}

	cpe := fmt.Sprintf("cpe:2.3:a:python:python:%s:*:*:*:*:*:*:*", version)
	licenses := retrieve.LookupLicenses(sourceURL, upstream.DefaultDecompress)
	purl := retrieve.GeneratePURL("python", version, sourceSHA256, sourceURL)

	return collections.TransformFuncWithError(getSuportedPlatformStackTargets(), func(platformTarget PlatformStackTarget) (versionology.Dependency, error) {
		fmt.Printf("Generating metadata for %s %s %s %s\n", platformTarget.os, platformTarget.arch, platformTarget.target, version)
		configMetadataDependency := cargo.ConfigMetadataDependency{
			CPE:             cpe,
			ID:              "python",
			Licenses:        licenses,
			Name:            "Python",
			PURL:            purl,
			Source:          sourceURL,
			SourceChecksum:  fmt.Sprintf("sha256:%s", sourceSHA256),
			Version:         version,
			DeprecationDate: eolDate,
			Stacks:          platformTarget.stacks,
			OS:              platformTarget.os,
			Arch:            platformTarget.arch,
		}

		if slices.Contains(platformTarget.stacks, "*") {
			configMetadataDependency.Checksum = configMetadataDependency.SourceChecksum
			configMetadataDependency.URI = configMetadataDependency.Source
		}
		return versionology.NewDependency(configMetadataDependency, platformTarget.target)
	})
}

func getSha256(sourceURL string, version string) (string, error) {
	resp, err := http.Get(sourceURL)
	if err != nil {
		return "", fmt.Errorf("failed to query url %q: %w", sourceURL, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to query url %q with: status code %d", sourceURL, resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp("", "python")
	if err != nil {
		return "", err
	}

	defer os.RemoveAll(tempDir)

	tarballPath := filepath.Join(tempDir, fmt.Sprintf("python-%s.tgz", version))
	err = os.WriteFile(tarballPath, body, os.ModePerm)
	if err != nil {
		return "", err
	}

	calculator := fs.NewChecksumCalculator()
	return calculator.Sum(tarballPath)
}

func getEOL(version *semver.Version) (*time.Time, error) {
	minorVersion := fmt.Sprintf("%d.%d", version.Major(), version.Minor())
	endpoint := fmt.Sprintf("https://endoflife.date/api/python/%s.json", minorVersion)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to query url %q: %w", endpoint, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query url %q with: status code %d", endpoint, resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type eolData struct {
		EolString string `json:"eol"`
	}

	d := eolData{}

	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal eol metadata: %w", err)
	}

	eol, err := time.Parse(time.DateOnly, d.EolString)
	if err != nil {
		return nil, fmt.Errorf("could not parse eol %q: %w", d.EolString, err)
	}

	return &eol, nil
}

func main() {
	retrieve.NewMetadata("python", getAllVersions, generateMetadata)
}
