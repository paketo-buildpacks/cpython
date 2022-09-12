package cpython_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/paketo-buildpacks/cpython"
	"github.com/paketo-buildpacks/cpython/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func testCPythonInstaller(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		sourcePath string
		layerPath  string
		workingDir string

		entry      packit.BuildpackPlanEntry
		dependency postal.Dependency

		configureProcess *fakes.Executable
		makeProcess      *fakes.Executable

		pythonInstaller cpython.PythonInstaller
	)

	it.Before(func() {
		var err error

		sourcePath, err = os.MkdirTemp("", "source")
		Expect(err).NotTo(HaveOccurred())

		layerPath, err = os.MkdirTemp("", "layer")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(layerPath, "bin"), 0755)).To(Succeed())

		workingDir, err = os.MkdirTemp("", "workingdir")
		Expect(err).NotTo(HaveOccurred())

		configureProcess = &fakes.Executable{}
		makeProcess = &fakes.Executable{}

		pythonInstaller = cpython.NewCPythonInstaller(configureProcess, makeProcess, scribe.NewEmitter(bytes.NewBuffer(nil)))
	})

	it.After(func() {
		Expect(os.RemoveAll(sourcePath)).To(Succeed())
		Expect(os.RemoveAll(layerPath)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Execute", func() {
		var (
			makeInvocationArgs [][]string
		)

		it.Before(func() {
			makeInvocationArgs = [][]string{}

			makeProcess.ExecuteCall.Stub = func(e pexec.Execution) error {
				makeInvocationArgs = append(makeInvocationArgs, e.Args)
				return nil
			}
		})

		it("runs installation", func() {
			err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(configureProcess.ExecuteCall.CallCount).To(Equal(1))
			Expect(configureProcess.ExecuteCall.Receives.Execution).To(MatchFields(IgnoreExtras, Fields{
				"Args": Equal([]string{
					"--enable-optimizations",
					"--with-ensurepip",
					fmt.Sprintf("--prefix=%s", layerPath),
				}),
			}))

			Expect(makeProcess.ExecuteCall.CallCount).To(Equal(2))

			Expect(makeInvocationArgs[0]).To(Equal([]string{
				"-j",
				fmt.Sprintf("%d", runtime.NumCPU()),
				`LDFLAGS="-Wl,--strip-all"`,
			}))

			Expect(makeInvocationArgs[1]).To(Equal([]string{
				"altinstall",
			}))
		})

		context("when configure flags are provided", func() {
			it.Before(func() {
				entry.Metadata = make(map[string]interface{})
				entry.Metadata["configure-flags"] = "--foo --bar=baz"
			})

			it("uses the provided flags instead of the default", func() {
				err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(configureProcess.ExecuteCall.CallCount).To(Equal(1))
				Expect(configureProcess.ExecuteCall.Receives.Execution).To(MatchFields(IgnoreExtras, Fields{
					"Args": Equal([]string{
						"--foo",
						"--bar=baz",
						fmt.Sprintf("--prefix=%s", layerPath),
					}),
				}))
			})
		})

		context("failure cases", func() {
			context("when changing to the source directory fails", func() {
				it.Before(func() {
					Expect(os.Chmod(sourcePath, 0000)).To(Succeed())
				})

				it("fails with error", func() {
					err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
					Expect(err).Should(MatchError(And(
						ContainSubstring(sourcePath),
						ContainSubstring("permission denied"),
					)))
				})
			})

			context("when invoking the configureProcess fails with error", func() {
				it.Before(func() {
					configureProcess.ExecuteCall.Returns.Error = errors.New("some configure error")
				})

				it("fails with error", func() {
					err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
					Expect(err).Should(MatchError("some configure error"))

				})
			})

			context("when invoking the make process (first time) fails with error", func() {
				it.Before(func() {
					makeProcess.ExecuteCall.Stub = nil
					makeProcess.ExecuteCall.Returns.Error = errors.New("some make error")
				})

				it("fails with error", func() {
					err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
					Expect(err).Should(MatchError("some make error"))

				})
			})

			context("when invoking the make process (second time) fails with error", func() {
				var (
					makeInvocationError error
				)

				it.Before(func() {
					makeInvocationError = errors.New("some make error (second invocation)")

					makeInvocationCount := 0
					makeProcess.ExecuteCall.Stub = func(_ pexec.Execution) error {
						makeInvocationCount++

						if makeInvocationCount == 1 {
							return nil
						}
						return makeInvocationError
					}
				})

				it("fails with error", func() {
					err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
					Expect(err).Should(MatchError("some make error (second invocation)"))

					Expect(makeProcess.ExecuteCall.CallCount).To(Equal(2))
				})
			})

			context("when changing to the layer bin directory fails", func() {
				it.Before(func() {
					Expect(os.Chmod(filepath.Join(layerPath, "bin"), 0000)).To(Succeed())
				})

				it("fails with error", func() {
					err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
					Expect(err).Should(MatchError(And(
						ContainSubstring(filepath.Join(layerPath, "bin")),
						ContainSubstring("permission denied"),
					)))
				})
			})

			context("when creating symlinks fails", func() {
				it.Before(func() {
					// 0500 permissions because we need Read (4) + Execute (1)
					// but not write (2) Execute permissions are required to cd
					// into the directory
					Expect(os.Chmod(filepath.Join(layerPath, "bin"), 0500)).To(Succeed())
				})

				it("fails with error", func() {
					err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
					Expect(err).Should(MatchError(ContainSubstring("symlink")))
				})
			})

			context("when changing back to the working directory fails", func() {
				it.Before(func() {
					Expect(os.Chmod(workingDir, 0000)).To(Succeed())
				})

				it("fails with error", func() {
					err := pythonInstaller.Install(sourcePath, workingDir, entry, dependency, layerPath)
					Expect(err).Should(MatchError(And(
						ContainSubstring(workingDir),
						ContainSubstring("permission denied"),
					)))
				})
			})
		})
	})
}
