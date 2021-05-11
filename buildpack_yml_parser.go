package cpython

import (
	"os"

	"gopkg.in/yaml.v2"
)

// BuildpackYMLParser parses the buildpack.yml file for Cpython-related
// configurations.
type BuildpackYMLParser struct{}

// NewBuildpackYMLParser creates a BuildpackYMLParser
func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

// ParseVersion decodes a given buildpack.yml file if it contains a Cpython or
// Python entry, and returns the the related version string.
func (p BuildpackYMLParser) ParseVersion(path string) (string, error) {
	var buildpack struct {
		Cpython struct {
			Version string `yaml:"version"`
		} `yaml:"cpython"`
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", err
	}

	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&buildpack)
	if err != nil {
		return "", err
	}

	return buildpack.Cpython.Version, nil
}
