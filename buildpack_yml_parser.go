package cpython

import (
	"os"

	"gopkg.in/yaml.v2"
)

type BuildpackYMLParser struct{}

func NewBuildpackYMLParser() BuildpackYMLParser {
	return BuildpackYMLParser{}
}

func (p BuildpackYMLParser) ParseVersion(path string) (string, error) {
	var buildpack struct {
		Cpython struct {
			Version string `yaml:"version"`
		} `yaml:"cpython"`

		// Restructure: Remove this after restructing is completed
		Python struct {
			Version string `yaml:"version"`
		} `yaml:"python"`
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

	// Restructure: Remove this after restructing is completed
	if buildpack.Cpython.Version != "" {
		return buildpack.Cpython.Version, nil
	}
	return buildpack.Python.Version, nil
}
