package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
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
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds with the defaults", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Cpython.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithCommand("python3 server.py").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).ToNot(HaveOccurred())

			Eventually(container).Should(BeAvailable())
			Eventually(container).Should(Serve(SatisfyAll(
				ContainSubstring("hello world"),
				ContainSubstring("PYTHONPYCACHEPREFIX=/home/cnb/.pycache"),
			)).OnPort(8080))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving CPython version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected CPython version \(using <unknown>\): `+strings.ReplaceAll(defaultVersion, ".", `\.`)),
			))
			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing CPython `+strings.ReplaceAll(defaultVersion, ".", `\.`)),
				MatchRegexp(`      Completed in \d+(\.?\d+)*`),
			))
			Expect(logs).To(ContainLines(
				"  Configuring build environment",
				fmt.Sprintf(`    PYTHONPATH          -> "/layers/%s/cpython"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
				`    PYTHONPYCACHEPREFIX -> "/tmp"`,
				"",
				"  Configuring launch environment",
				fmt.Sprintf(`    PYTHONPATH -> "/layers/%s/cpython"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
			))
		})

		it("should have a runnable 'python' on the path", func() {
			var err error
			var logs fmt.Stringer
			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Cpython.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithCommand("python server.py").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).ToNot(HaveOccurred())

			Eventually(container).Should(BeAvailable())
			Eventually(container).Should(Serve(SatisfyAll(
				ContainSubstring("hello world"),
				ContainSubstring("PYTHONPYCACHEPREFIX=/home/cnb/.pycache"),
			)).OnPort(8080))
		})

		context("validating SBOM", func() {
			var (
				container2 occam.Container
				sbomDir    string
			)

			it.Before(func() {
				var err error
				sbomDir, err = os.MkdirTemp("", "sbom")
				Expect(err).NotTo(HaveOccurred())
				Expect(os.Chmod(sbomDir, os.ModePerm)).To(Succeed())
			})

			it.After(func() {
				Expect(docker.Container.Remove.Execute(container2.ID)).To(Succeed())
				Expect(os.RemoveAll(sbomDir)).To(Succeed())
			})

			it("writes SBOM files to the layer and label metadata", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.Cpython.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithEnv(map[string]string{
						"BP_LOG_LEVEL": "DEBUG",
					}).
					WithSBOMOutputDir(sbomDir).
					Execute(name, source)
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
					fmt.Sprintf("  Generating SBOM for /layers/%s/cpython", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					MatchRegexp(`      Completed in \d+(\.?\d+)*`),
				))
				Expect(logs).To(ContainLines(
					"  Writing SBOM in the following format(s):",
					"    application/vnd.cyclonedx+json",
					"    application/spdx+json",
					"    application/vnd.syft+json",
				))

				// check that legacy SBOM is included via metadata
				container2, err = docker.Container.Run.
					WithCommand("cat /layers/sbom/launch/sbom.legacy.json").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container2.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(ContainSubstring(`"name":"CPython"`))

				// check that all required SBOM files are present
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "cpython", "sbom.cdx.json")).To(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "cpython", "sbom.spdx.json")).To(BeARegularFile())
				Expect(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "cpython", "sbom.syft.json")).To(BeARegularFile())

				// check an SBOM file to make sure it has an entry for cpython
				contents, err := os.ReadFile(filepath.Join(sbomDir, "sbom", "launch", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"), "cpython", "sbom.cdx.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(ContainSubstring(`"name": "CPython"`))
			})
		})

		context("when the BP_CPYTHON_VERSION environment variable is set", func() {
			it("builds with the requested version of python", func() {
				var (
					err               error
					logs              fmt.Stringer
					nonDefaultVersion string
				)

				for _, dependency := range dependenciesForStack() {
					if dependency.Version != defaultVersion {
						nonDefaultVersion = dependency.Version
						break
					}
				}

				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.Cpython.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithEnv(map[string]string{
						"BP_CPYTHON_VERSION": nonDefaultVersion,
					}).
					Execute(name, source)
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
					`      BP_CPYTHON_VERSION -> "`+nonDefaultVersion+`"`,
					`      <unknown>          -> ""`,
				))

				Expect(logs).To(ContainLines(
					MatchRegexp(`   Selected CPython version \(using BP_CPYTHON_VERSION\): ` + strings.ReplaceAll(nonDefaultVersion, ".", `\.`)),
				))

				Expect(logs).To(ContainLines(
					"  Executing build process",
					"    Installing CPython "+nonDefaultVersion,
					MatchRegexp(`      Completed in \d+(\.?\d+)*`),
				))

				Expect(logs).To(ContainLines(
					"  Configuring build environment",
					fmt.Sprintf(`    PYTHONPATH          -> "/layers/%s/cpython"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
					`    PYTHONPYCACHEPREFIX -> "/tmp"`,
					"",
					"  Configuring launch environment",
					fmt.Sprintf(`    PYTHONPATH -> "/layers/%s/cpython"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
				))
			})
		})

		context("when the BP_CPYTHON_RM_SETUPTOOLS environment variable is set", func() {
			var container2 occam.Container

			it.After(func() {
				Expect(docker.Container.Remove.Execute(container2.ID)).To(Succeed())
			})

			it("builds with the defaults and setuptools is not installed", func() {
				var err error
				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						settings.Buildpacks.Cpython.Online,
						settings.Buildpacks.BuildPlan.Online,
					).
					WithEnv(map[string]string{
						"BP_CPYTHON_RM_SETUPTOOLS": "value-is-ignored",
					}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithCommand("python3 server.py").
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).ToNot(HaveOccurred())

				Eventually(container).Should(BeAvailable())
				Eventually(container).Should(Serve(ContainSubstring("hello world")).OnPort(8080))

				// check that setuptools is not installed.
				// since we cannot get the return code, echo a string that we can search for if the command fails
				container2, err = docker.Container.Run.
					WithCommand("python3 -m pip show setuptools || echo not-installed").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					cLogs, err := docker.Container.Logs.Execute(container2.ID)
					Expect(err).NotTo(HaveOccurred())
					return cLogs.String()
				}).Should(ContainSubstring("not-installed"))
			})
		})
	})
}
