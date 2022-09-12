package integration_test

import (
	"os"
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

		name   string
		source string
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

		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("when an app is rebuilt and does not change", func() {
		it("reuses a layer from a previous build", func() {
			var (
				err         error
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			build := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Cpython.Online,
					settings.Buildpacks.BuildPlan.Online,
				)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			firstImage, _, err = build.
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("cpython"))

			firstContainer, err = docker.Container.Run.
				WithCommand("python3 server.py").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(firstImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			// Second pack build
			secondImage, _, err = build.
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("cpython"))

			secondContainer, err = docker.Container.Run.
				WithCommand("python3 server.py").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(secondImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())
			Eventually(secondContainer).Should(Serve(ContainSubstring("hello world")).OnPort(8080))

			Expect(secondImage.Buildpacks[0].Layers["cpython"].SHA).To(Equal(firstImage.Buildpacks[0].Layers["cpython"].SHA))
		})
	})

	context("when an app is rebuilt and there is a change", func() {
		it("rebuilds the layer", func() {
			var (
				err         error
				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container

				otherVersion string
			)

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())

			firstImage, _, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Cpython.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_CPYTHON_VERSION": defaultVersion}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[firstImage.ID] = struct{}{}

			Expect(firstImage.Buildpacks).To(HaveLen(2))
			Expect(firstImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(firstImage.Buildpacks[0].Layers).To(HaveKey("cpython"))

			firstContainer, err = docker.Container.Run.
				WithCommand("python3 server.py").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(firstImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			Eventually(firstContainer).Should(BeAvailable())

			for _, dependency := range buildpackInfo.Metadata.Dependencies {
				if dependency.Version != defaultVersion {
					otherVersion = dependency.Version
					break
				}
			}

			secondImage, _, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.Cpython.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_CPYTHON_VERSION": otherVersion}).
				Execute(name, source)
			Expect(err).NotTo(HaveOccurred())

			imageIDs[secondImage.ID] = struct{}{}

			Expect(secondImage.Buildpacks).To(HaveLen(2))
			Expect(secondImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers).To(HaveKey("cpython"))

			secondContainer, err = docker.Container.Run.
				WithCommand("python3 server.py").
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(secondImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(secondContainer).Should(BeAvailable())
			Eventually(secondContainer).Should(Serve(ContainSubstring("hello world")).OnPort(8080))

			Expect(secondImage.Buildpacks[0].Layers["cpython"].SHA).NotTo(Equal(firstImage.Buildpacks[0].Layers["cpython"].SHA))
		})
	})
}
