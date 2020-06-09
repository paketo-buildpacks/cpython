package pythonruntime

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

	return buildpack.Python.Version, nil
}
