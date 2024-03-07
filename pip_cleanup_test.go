package cpython_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/paketo-buildpacks/cpython"
	"github.com/paketo-buildpacks/cpython/fakes"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPipCleanup(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layerPath     string
		pythonProcess *fakes.Executable
		pipCleanup    cpython.PythonPipCleanup
	)

	it.Before(func() {
		pythonProcess = &fakes.Executable{}

		pipCleanup = cpython.NewPipCleanup(pythonProcess, scribe.NewEmitter(bytes.NewBuffer(nil)))
	})

	context("Execute", func() {
		var (
			pythonInvocationArgs [][]string
			packages             []string
		)

		it.Before(func() {
			pythonInvocationArgs = [][]string{}
			packages = []string{}

			pythonProcess.ExecuteCall.Stub = func(e pexec.Execution) error {
				pythonInvocationArgs = append(pythonInvocationArgs, e.Args)
				return nil
			}
		})

		context("when packages are installed", func() {
			it.Before(func() {
				packages = []string{"somepkg"}
			})

			it("uninstalls packages", func() {
				err := pipCleanup.Cleanup(packages, layerPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(pythonProcess.ExecuteCall.CallCount).To(Equal(3))
				Expect(pythonInvocationArgs[0]).To(Equal([]string{"-m", "pip", "--version"}))
				Expect(pythonInvocationArgs[1]).To(Equal([]string{"-m", "pip", "show", "-q", packages[0]}))
				Expect(pythonInvocationArgs[2]).To(Equal([]string{"-m", "pip", "uninstall", "-y", packages[0]}))
			})
		})

		context("when packages are not installed", func() {
			it.Before(func() {
				packages = []string{"somepkg-that-does-not-exist"}

				pythonProcess.ExecuteCall.Stub = func(e pexec.Execution) error {
					pythonInvocationArgs = append(pythonInvocationArgs, e.Args)
					if e.Args[2] == "show" {
						return errors.New("pip package not found")
					}
					return nil
				}
			})

			it("does not uninstall packages", func() {
				err := pipCleanup.Cleanup(packages, layerPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(pythonProcess.ExecuteCall.CallCount).To(Equal(2))
				Expect(pythonInvocationArgs[0]).To(Equal([]string{"-m", "pip", "--version"}))
				Expect(pythonInvocationArgs[1]).To(Equal([]string{"-m", "pip", "show", "-q", packages[0]}))
			})
		})

		context("failure cases", func() {
			context("when pip version returns an error", func() {
				it.Before(func() {
					packages = []string{"somepkg"}

					pythonProcess.ExecuteCall.Stub = func(e pexec.Execution) error {
						pythonInvocationArgs = append(pythonInvocationArgs, e.Args)
						return errors.New("pip is broken")
					}
				})

				it("fails with error", func() {
					err := pipCleanup.Cleanup(packages, layerPath)
					Expect(err).Should(MatchError(And(
						ContainSubstring("pip is broken"),
					)))
					Expect(pythonProcess.ExecuteCall.CallCount).To(Equal(1))
					Expect(pythonInvocationArgs[0]).To(Equal([]string{"-m", "pip", "--version"}))
				})
			})

			context("when pip uninstall returns an error", func() {
				it.Before(func() {
					packages = []string{"somepkg"}

					pythonProcess.ExecuteCall.Stub = func(e pexec.Execution) error {
						pythonInvocationArgs = append(pythonInvocationArgs, e.Args)
						if e.Args[2] == "uninstall" {
							return errors.New("failed to uninstall pip package")
						}
						return nil
					}
				})

				it("fails with error", func() {
					err := pipCleanup.Cleanup(packages, layerPath)
					Expect(err).Should(MatchError(And(
						ContainSubstring("failed to uninstall pip package"),
					)))
					Expect(pythonProcess.ExecuteCall.CallCount).To(Equal(3))
					Expect(pythonInvocationArgs[0]).To(Equal([]string{"-m", "pip", "--version"}))
					Expect(pythonInvocationArgs[1]).To(Equal([]string{"-m", "pip", "show", "-q", packages[0]}))
					Expect(pythonInvocationArgs[2]).To(Equal([]string{"-m", "pip", "uninstall", "-y", packages[0]}))
				})
			})
		})
	})
}
