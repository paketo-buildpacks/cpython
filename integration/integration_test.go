package integration

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	bp string
)

func TestIntegration(t *testing.T) {
	RegisterTestingT(t)
	root, err := dagger.FindBPRoot()
	Expect(err).ToNot(HaveOccurred())
	bp, err = dagger.PackageBuildpack(root)
	Expect(err).NotTo(HaveOccurred())
	defer func() {
		dagger.DeleteBuildpack(bp)
	}()

	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, _ spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("can run a python app", func() {
		app, err := dagger.PackBuild(filepath.Join("testdata", "app"), bp)
		Expect(err).ToNot(HaveOccurred())
		app.Memory = "128m"
		defer app.Destroy()

		Expect(app.StartWithCommand("python3 server.py")).To(Succeed())

		body, _, err := app.HTTPGet("/")
		Expect(err).NotTo(HaveOccurred())
		Expect(body).To(ContainSubstring("hello world"))
	})
}
