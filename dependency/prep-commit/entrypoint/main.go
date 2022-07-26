package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

func main() {
	var (
		uri               string
		sha256            string
		version           string
		buildpackTomlPath string
		upstreamUri       string
		upstreamSha256    string
		stack             string
		verbose           bool
	)

	flag.StringVar(&uri, "uri", "", "URI of the dependency")
	flag.StringVar(&sha256, "sha256", "", "SHA256 of the dependency")
	flag.StringVar(&version, "version", "", "version of the dependency")
	flag.StringVar(&buildpackTomlPath, "buildpack-toml-path", "", "path to buildpack.toml")
	flag.StringVar(&upstreamUri, "upstream-uri", "", "URI of the upstream dependency")
	flag.StringVar(&upstreamSha256, "upstream-sha256", "", "SHA256 of the upstream dependency")
	flag.StringVar(&stack, "stack", "", "stack id")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")
	flag.Parse()

	if verbose {
		fmt.Printf("uri: %s\n", uri)
		fmt.Printf("sha256: %s\n", sha256)
		fmt.Printf("buildpackTomlPath: %s\n", buildpackTomlPath)
		fmt.Printf("version: %s\n", version)
		fmt.Printf("upstreamUri: %s\n", upstreamUri)
		fmt.Printf("upstreamSha256: %s\n", upstreamSha256)
		fmt.Printf("stack: %s\n", stack)
		fmt.Printf("verbose: %t\n", verbose)
	}

	config := parseBuildpackToml(buildpackTomlPath)

	newDependency := cargo.ConfigMetadataDependency{}
	newDependency.CPE = fmt.Sprintf("cpe:2.3:a:python:python:%s:*:*:*:*:*:*:*", version)
	newDependency.ID = "python"
	newDependency.Licenses = []interface{}{
		"0BSD",
		"CNRI-Python-GPL-Compatible",
	}
	newDependency.Name = "Python"
	newDependency.PURL = fmt.Sprintf("pkg:generic/python@%s?checksum=%s&download_url=%s",
		version,
		upstreamSha256,
		upstreamUri)
	newDependency.SHA256 = sha256
	newDependency.Source = upstreamUri
	newDependency.SourceSHA256 = upstreamSha256
	newDependency.Stacks = []string{
		stack,
	}
	newDependency.URI = uri
	newDependency.Version = version

	config.Metadata.Dependencies = append(config.Metadata.Dependencies, newDependency)

	file, err := os.OpenFile(buildpackTomlPath, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Errorf("failed to open buildpack config file: %w", err))
	}
	defer file.Close()

	err = cargo.EncodeConfig(file, config)
	if err != nil {
		panic(fmt.Errorf("failed to write buildpack config: %w", err))
	}
}

func parseBuildpackToml(buildpackTomlPath string) cargo.Config {
	configParser := cargo.NewBuildpackParser()
	config, err := configParser.Parse(buildpackTomlPath)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s: %s", buildpackTomlPath, err))
	}
	return config
}
