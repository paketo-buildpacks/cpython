package cpython_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPython(t *testing.T) {
	suite := spec.New("python", spec.Report(report.Terminal{}), spec.Sequential())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("CPythonInstaller", testCPythonInstaller)
	suite("PipCleanup", testPipCleanup)
	suite.Run(t)
}
