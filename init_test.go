package cpython_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPython(t *testing.T) {
	suite := spec.New("python", spec.Report(report.Terminal{}))
	suite("Build", testBuild)
	suite("BuildpackYMLParser", testBuildpackYMLParser)
	suite("Detect", testDetect)
	suite("LogEmitter", testLogEmitter)
	suite("PlanEntryResolver", testPlanEntryResolver)
	suite("PlanRefinery", testPlanRefinery)
	suite.Run(t)
}
