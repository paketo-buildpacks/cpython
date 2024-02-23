package main

import (
	"bytes"
	_ "embed"
	"os"
	"text/template"
)

type URI struct {
	Prefix string
}

//go:embed buildpack.toml
var BuildpackToml string

func main() {
	paketoURI := URI{"https://artifacts.paketo.io/python/"}
	tmpl, err := template.New("buildpack.toml").Parse(BuildpackToml)
	if err != nil {
		panic("failed to parse buildpack.toml")
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, paketoURI)
	if err != nil {
		panic("failed to execute template")
	}

	err = os.WriteFile("buildpack.toml", b.Bytes(), 0600)
	if err != nil {
		panic("failed to write buildpack.toml")
	}
}
