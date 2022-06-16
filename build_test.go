package cpython_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/cpython"
	"github.com/paketo-buildpacks/cpython/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"

	//nolint Ignore SA1019, informed usage of deprecated package
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir         string
		cnbDir            string
		clock             chronos.Clock
		dependencyManager *fakes.DependencyManager
		sbomGenerator     *fakes.SBOMGenerator
		buffer            *bytes.Buffer
		logEmitter        scribe.Emitter

		build        packit.BuildFunc
		buildContext packit.BuildContext
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		clock = chronos.DefaultClock

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			// Dependecy is called python not cpython
			ID:      "python",
			Name:    "python-dependency-name",
			SHA256:  "python-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "python-dependency-uri",
			Version: "python-dependency-version",
		}

		// Legacy SBOM
		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "cpython",
				Metadata: paketosbom.BOMMetadata{
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "cpython-dependency-sha",
					},
					URI:     "cpython-dependency-uri",
					Version: "cpython-dependency-version",
				},
			},
		}

		// Syft SBOM
		sbomGenerator = &fakes.SBOMGenerator{}
		sbomGenerator.GenerateFromDependencyCall.Returns.SBOM = sbom.SBOM{}

		buffer = bytes.NewBuffer(nil)
		logEmitter = scribe.NewEmitter(buffer)

		build = cpython.Build(dependencyManager, sbomGenerator, logEmitter, clock)

		buildContext = packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "cpython",
				Version:     "some-version",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			CNBPath: cnbDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{Name: "cpython"},
				},
			},
			Platform: packit.Platform{Path: "platform"},
			Layers:   packit.Layers{Path: layersDir},
			Stack:    "some-stack",
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs python", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("cpython"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "cpython")))

		Expect(layer.SharedEnv).To(Equal(packit.Environment{
			"PYTHONPATH.default": filepath.Join(layersDir, "cpython"),
		}))
		Expect(layer.BuildEnv).To(BeEmpty())
		Expect(layer.LaunchEnv).To(BeEmpty())
		Expect(layer.ProcessLaunchEnv).To(BeEmpty())

		Expect(layer.Build).To(BeFalse())
		Expect(layer.Launch).To(BeFalse())
		Expect(layer.Cache).To(BeFalse())

		Expect(layer.Metadata).To(Equal(map[string]interface{}{
			"dependency-sha": "python-dependency-sha",
		}))

		Expect(layer.ExecD).To(Equal([]string{
			filepath.Join(cnbDir, "bin", "env"),
		}))

		Expect(layer.SBOM.Formats()).To(Equal([]packit.SBOMFormat{
			{
				Extension: sbom.Format(sbom.CycloneDXFormat).Extension(),
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.CycloneDXFormat),
			},
			{
				Extension: sbom.Format(sbom.SPDXFormat).Extension(),
				Content:   sbom.NewFormattedReader(sbom.SBOM{}, sbom.SPDXFormat),
			},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		// Dependecy is called python not cpython
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("python"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal(""))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "cpython",
			Name:    "CPython",
			SHA256:  "python-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "python-dependency-uri",
			Version: "python-dependency-version",
		}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.DestinationPath).To(Equal(filepath.Join(layersDir, "cpython")))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("platform"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
			{
				ID:      "cpython",
				Name:    "CPython",
				SHA256:  "python-dependency-sha",
				Stacks:  []string{"some-stack"},
				URI:     "python-dependency-uri",
				Version: "python-dependency-version",
			},
		}))

		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "cpython",
			Name:    "CPython",
			SHA256:  "python-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "python-dependency-uri",
			Version: "python-dependency-version",
		}))
		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dir).To(Equal(filepath.Join(layersDir, "cpython")))

		Expect(buffer.String()).To(ContainSubstring("cpython some-version"))
		Expect(buffer.String()).To(ContainSubstring("Resolving CPython version"))
		Expect(buffer.String()).To(ContainSubstring("Selected CPython version (using <unknown>): python-dependency-version"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Installing CPython python-dependency-version"))
		Expect(buffer.String()).To(ContainSubstring("Completed in"))
	})

	context("when the plan entry requires the dependency during the build and launch phases", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{"build": true, "launch": true}
		})

		it("makes the layer available in those phases", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("cpython"))

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeTrue())
			Expect(layer.Cache).To(BeTrue())

			Expect(result.Build.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "cpython",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "cpython-dependency-sha",
							},
							URI:     "cpython-dependency-uri",
							Version: "cpython-dependency-version",
						},
					},
				},
			))

			Expect(result.Launch.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "cpython",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "cpython-dependency-sha",
							},
							URI:     "cpython-dependency-uri",
							Version: "cpython-dependency-version",
						},
					},
				},
			))
		})
	})

	context("when the cached SHA matches the dependency SHA", func() {
		it.Before(func() {
			err := os.WriteFile(filepath.Join(layersDir, "cpython.toml"), []byte("[metadata]\ndependency-sha = \"python-dependency-sha\"\n"), 0600)
			Expect(err).NotTo(HaveOccurred())

			buildContext.Plan.Entries[0].Metadata = map[string]interface{}{"build": true}
		})

		it("reuses the existing layer", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("cpython"))

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeFalse())
			Expect(layer.Cache).To(BeTrue())

			Expect(result.Build.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "cpython",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "cpython-dependency-sha",
							},
							URI:     "cpython-dependency-uri",
							Version: "cpython-dependency-version",
						},
					},
				},
			))

			Expect(dependencyManager.DeliverCall.CallCount).To(Equal(0))
			Expect(sbomGenerator.GenerateFromDependencyCall.CallCount).To(Equal(0))

		})
	})

	context("failure cases", func() {
		context("when the dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when generating the SBOM returns an error", func() {
			it.Before(func() {
				buildContext.BuildpackInfo.SBOMFormats = []string{"random-format"}
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(`unsupported SBOM format: 'random-format'`))
			})
		})

		context("when formatting the SBOM returns an error", func() {
			it.Before(func() {
				sbomGenerator.GenerateFromDependencyCall.Returns.Error = errors.New("failed to generate SBOM")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to generate SBOM")))
			})
		})

		context("when the python layer cannot be retrieved", func() {
			it.Before(func() {
				err := os.WriteFile(filepath.Join(layersDir, "cpython.toml"), nil, 0000)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata")))
			})
		})

		context("when the python layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, "cpython", "something"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(layersDir, "cpython"), 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layersDir, "cpython"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("could not remove file")))
			})
		})

		context("when the dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.DeliverCall.Returns.Error = errors.New("failed to install dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})
	})
}
