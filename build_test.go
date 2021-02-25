package cpython_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	cpython "github.com/paketo-community/cpython"
	"github.com/paketo-community/cpython/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir         string
		cnbDir            string
		timeStamp         time.Time
		clock             chronos.Clock
		entryResolver     *fakes.EntryResolver
		dependencyManager *fakes.DependencyManager
		planRefinery      *fakes.PlanRefinery
		buffer            *bytes.Buffer

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		timeStamp = time.Now()
		clock = chronos.NewClock(func() time.Time {
			return timeStamp
		})

		entryResolver = &fakes.EntryResolver{}
		entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "python",
		}

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:      "python",
			Name:    "python-dependency-name",
			SHA256:  "python-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "python-dependency-uri",
			Version: "python-dependency-version",
		}

		planRefinery = &fakes.PlanRefinery{}
		planRefinery.BillOfMaterialsCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
			Name: "python",
			Metadata: map[string]interface{}{
				"name":    "python-dependency-name",
				"sha256":  "python-dependency-sha",
				"stacks":  []string{"some-stack"},
				"uri":     "python-dependency-uri",
				"version": "python-dependency-version",
			},
		}

		buffer = bytes.NewBuffer(nil)
		logEmitter := cpython.NewLogEmitter(buffer)

		build = cpython.Build(entryResolver, dependencyManager, planRefinery, logEmitter, clock)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs python", func() {
		result, err := build(packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			CNBPath: cnbDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{Name: "python"},
				},
			},
			Layers: packit.Layers{Path: layersDir},
			Stack:  "some-stack",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "python",
						Metadata: map[string]interface{}{
							"name":    "python-dependency-name",
							"sha256":  "python-dependency-sha",
							"stacks":  []string{"some-stack"},
							"uri":     "python-dependency-uri",
							"version": "python-dependency-version",
						},
					},
				},
			},
			Layers: []packit.Layer{
				{
					Name: "python",
					Path: filepath.Join(layersDir, "python"),
					SharedEnv: packit.Environment{
						"PYTHONPATH.override": filepath.Join(layersDir, "python"),
					},
					BuildEnv:  packit.Environment{},
					LaunchEnv: packit.Environment{},
					Build:     false,
					Launch:    false,
					Cache:     false,
					Metadata: map[string]interface{}{
						"dependency-sha": "python-dependency-sha",
						"built_at":       timeStamp.Format(time.RFC3339Nano),
					},
				},
			},
		}))

		Expect(entryResolver.ResolveCall.Receives.BuildpackPlanEntrySlice).To(Equal([]packit.BuildpackPlanEntry{
			{Name: "python"},
		}))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("python"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal(""))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.InstallCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "python",
			Name:    "python-dependency-name",
			SHA256:  "python-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "python-dependency-uri",
			Version: "python-dependency-version",
		}))
		Expect(dependencyManager.InstallCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.InstallCall.Receives.LayerPath).To(Equal(filepath.Join(layersDir, "python")))

		Expect(planRefinery.BillOfMaterialsCall.Receives.Dependency).To(Equal(postal.Dependency{
			ID:      "python",
			Name:    "python-dependency-name",
			SHA256:  "python-dependency-sha",
			Stacks:  []string{"some-stack"},
			URI:     "python-dependency-uri",
			Version: "python-dependency-version",
		}))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Resolving Python version"))
		Expect(buffer.String()).To(ContainSubstring("Selected python-dependency-name version (using <unknown>): python-dependency-version"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Installing Python python-dependency-version"))
		Expect(buffer.String()).To(ContainSubstring("Completed in"))
	})

	context("when the plan entry requires the dependency during the build and launch phases", func() {
		it.Before(func() {
			entryResolver.ResolveCall.Returns.BuildpackPlanEntry = packit.BuildpackPlanEntry{
				Name: "python",
				Metadata: map[string]interface{}{
					"build":  true,
					"launch": true,
				},
			}
		})

		it("makes the layer available in those phases", func() {
			result, err := build(packit.BuildContext{
				CNBPath: cnbDir,
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "python",
							Metadata: map[string]interface{}{
								"build":  true,
								"launch": true,
							},
						},
					},
				},
				Layers: packit.Layers{Path: layersDir},
				Stack:  "some-stack",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: "python",
							Metadata: map[string]interface{}{
								"name":    "python-dependency-name",
								"sha256":  "python-dependency-sha",
								"stacks":  []string{"some-stack"},
								"uri":     "python-dependency-uri",
								"version": "python-dependency-version",
							},
						},
					},
				},
				Layers: []packit.Layer{
					{
						Name: "python",
						Path: filepath.Join(layersDir, "python"),
						SharedEnv: packit.Environment{
							"PYTHONPATH.override": filepath.Join(layersDir, "python"),
						},
						BuildEnv:  packit.Environment{},
						LaunchEnv: packit.Environment{},
						Build:     true,
						Launch:    true,
						Cache:     true,
						Metadata: map[string]interface{}{
							"dependency-sha": "python-dependency-sha",
							"built_at":       timeStamp.Format(time.RFC3339Nano),
						},
					},
				},
			}))
		})
	})

	context("failure cases", func() {
		context("when the dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "python"},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})

		context("when the python layer cannot be retrieved", func() {
			it.Before(func() {
				err := ioutil.WriteFile(filepath.Join(layersDir, "python.toml"), nil, 0000)
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "python"},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse layer content metadata")))
			})
		})

		context("when the python layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, "python", "something"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(layersDir, "python"), 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layersDir, "python"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "python"},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError(ContainSubstring("could not remove file")))
			})
		})

		context("when the dependency cannot be installed", func() {
			it.Before(func() {
				dependencyManager.InstallCall.Returns.Error = errors.New("failed to install dependency")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath: cnbDir,
					Plan: packit.BuildpackPlan{
						Entries: []packit.BuildpackPlanEntry{
							{Name: "python"},
						},
					},
					Layers: packit.Layers{Path: layersDir},
					Stack:  "some-stack",
				})
				Expect(err).To(MatchError("failed to install dependency"))
			})
		})
	})
}
