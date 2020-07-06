package integration_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testLayerReuse(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		docker occam.Docker
		pack   occam.Pack

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}
		name         string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		docker = occam.NewDocker()
		pack = occam.NewPack()
		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}

	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
	})

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			firstImage, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal("paketo-community/python-runtime"))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("python"))

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Python Runtime Buildpack %s", buildpackVersion),
				"  Resolving Python version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Python version \(using <unknown>\): 3\.8\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Python 3\.8\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))

			firstContainer, err = docker.Container.Run.WithCommand("python3 server.py").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal("paketo-community/python-runtime"))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("python"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Python Runtime Buildpack %s", buildpackVersion),
				"  Resolving Python version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
				"",
				MatchRegexp(`    Selected Python version \(using <unknown>\): 3\.8\.\d+`),
				"",
				"  Reusing cached layer /layers/paketo-community_python-runtime/python",
			))

			secondContainer, err = docker.Container.Run.WithCommand("python3 server.py").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("hello world"))

			Expect(secondImage.Buildpacks[0].Layers["python"].Metadata["built_at"]).To(Equal(firstImage.Buildpacks[0].Layers["python"].Metadata["built_at"]))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		it("rebuilds the layer", func() {
			var (
				err         error
				logs        fmt.Stringer
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			firstImage, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "buildpack_yml_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal("paketo-community/python-runtime"))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("python"))

			buildpackVersion, err := GetGitVersion()
			Expect(err).ToNot(HaveOccurred())

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Python Runtime Buildpack %s", buildpackVersion),
				"  Resolving Python version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"~3\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected Python version \(using buildpack.yml\): 3\.8\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Python 3\.8\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))

			firstContainer, err = docker.Container.Run.WithCommand("python3 server.py").Execute(firstImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, logs, err = pack.WithNoColor().Build.
				WithNoPull().
				WithBuildpacks(buildpack, buildPlanBuildpack).
				Execute(name, filepath.Join("testdata", "different_version_buildpack_yml_app"))
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal("paketo-community/python-runtime"))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("python"))

			Expect(logs).To(ContainLines(
				fmt.Sprintf("Python Runtime Buildpack %s", buildpackVersion),
				"  Resolving Python version",
				"    Candidate version sources (in priority order):",
				"      buildpack.yml -> \"3.7.*\"",
				"      <unknown>     -> \"\"",
				"",
				MatchRegexp(`    Selected Python version \(using buildpack.yml\): 3\.7\.\d+`),
				"",
				"  Executing build process",
				MatchRegexp(`    Installing Python 3\.7\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))

			secondContainer, err = docker.Container.Run.WithCommand("python3 server.py").Execute(secondImage.ID)
			Expect(err).NotTo(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())

			response, err := http.Get(fmt.Sprintf("http://localhost:%s", secondContainer.HostPort()))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))

			content, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(content).To(ContainSubstring("hello world"))

			Expect(secondImage.Buildpacks[0].Layers["python"].Metadata["built_at"]).NotTo(Equal(firstImage.Buildpacks[0].Layers["python"].Metadata["built_at"]))
		})
	})
}
