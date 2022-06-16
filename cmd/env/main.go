package main

import (
	"log"
	"os"

	"github.com/paketo-buildpacks/cpython/cmd/env/internal"
)

// env will set environment variables that are dynamically defined at runtime
func main() {
	err := internal.Run(os.Environ(), os.NewFile(3, "/dev/fd/3"))
	if err != nil {
		log.Fatal(err)
	}
}
