package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

func main() {
	var (
		minVersionString  string
		buildpackTomlPath string
		stack             string
		verbose           bool
	)

	flag.StringVar(&minVersionString, "min-version", "", "minimum version to check")
	flag.StringVar(&buildpackTomlPath, "buildpack-toml-path", "", "path to buildpack.toml")
	flag.StringVar(&stack, "stack", "", "stack id")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")
	flag.Parse()

	minVersion := semver.MustParse(minVersionString)

	if verbose {
		fmt.Printf("minVersion: %s\n", minVersion.String())
		fmt.Printf("buildpackTomlPath: %s\n", buildpackTomlPath)
		fmt.Printf("stack: %s\n", stack)
		fmt.Printf("verbose: %t\n", verbose)
	}

	versions, err := GetAllVersionRefs()
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	versionsAtLeastMin := make([]*semver.Version, 0)

	for _, versionString := range versions {
		version := semver.MustParse(versionString)

		if version.GreaterThan(minVersion) || version.Equal(minVersion) {
			versionsAtLeastMin = append(versionsAtLeastMin, version)
		}
	}

	GetNewVersions("python", "Python", stack, buildpackTomlPath, versionsAtLeastMin)

	newVersionsJSON, err := json.Marshal(versionsAtLeastMin)
	if err != nil {
		fmt.Printf("failed to marshal new versions: %e", err)
		os.Exit(1)
	}

	fmt.Println(string(newVersionsJSON))
}

func GetAllVersionRefs() ([]string, error) {
	webClient := NewWebClient()
	body, err := webClient.Get("https://www.python.org/downloads/")
	if err != nil {
		return nil, fmt.Errorf("could not get python downloads: %w", err)
	}

	re := regexp.MustCompile(`release-number.*Python ([\d]+\.[\d]+\.[\d]+)`)

	var versions []string
	for _, line := range strings.Split(string(body), "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			versions = append(versions, matches[1])
		}
	}

	return versions, nil
}

type WebClient struct {
	httpClient *http.Client
}

func NewWebClient() WebClient {
	return WebClient{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

type RequestOption func(r *http.Request)

func (w WebClient) Get(url string) ([]byte, error) {
	responseBody, err := w.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(responseBody.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}
	return body, nil
}
