package integration_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testBuildpackYAML(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		})

		it("builds with the settings in buildpack.yml AND logs a deprecation warning", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Cpython.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, filepath.Join("testdata", "buildpack_yml_app"))
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithCommand("python3 server.py").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).ToNot(HaveOccurred())

			Eventually(container).Should(BeAvailable())
			Eventually(container).Should(Serve(ContainSubstring("hello world")).OnPort(8080))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving CPython version",
				"    Candidate version sources (in priority order):",
				`      buildpack.yml -> "~3"`,
				`      <unknown>     -> ""`,
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected CPython version \(using buildpack.yml\): 3\.\d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`    WARNING\: Setting the CPython version through buildpack\.yml is deprecated and will be removed in %s v2.0.0.`, buildpackInfo.Buildpack.Name)),
				"    Please specify the version through the $BP_CPYTHON_VERSION environment variable instead. See docs for more information.",
			))
			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing CPython 3\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				"  Configuring environment",
				MatchRegexp(fmt.Sprintf(`    PYTHONPATH -> "/layers/%s/cpython"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
			))
		})
	})
}
