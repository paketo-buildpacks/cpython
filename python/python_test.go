package python_test

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildpackplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/cloudfoundry/python-runtime-cnb/python"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPython(t *testing.T) {
	RegisterTestingT(t)
	spec.Run(t, "Python", testPython, spec.Report(report.Terminal{}))
}

func testPython(t *testing.T, when spec.G, it spec.S) {
	var stubPythonFixture = filepath.Join("testdata", "stub-python.tar.gz")

	when("NewContributor", func() {
		it("returns true if python is in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddPlan(buildpackplan.Plan{Name: python.Dependency})
			f.AddDependency(python.Dependency, stubPythonFixture)

			_, willContribute, err := python.NewContributor(f.Build)
			if err != nil {
				t.Error("Could not create contributor")
			}

			if diff := cmp.Diff(willContribute, true); diff != "" {
				t.Error("Python should contribute to the build plan")
			}

		})

		it("returns false if python is not in the build plan", func() {
			f := test.NewBuildFactory(t)

			_, willContribute, err := python.NewContributor(f.Build)
			if err != nil {
				t.Error("Could not create contributor")
			}

			if diff := cmp.Diff(willContribute, false); diff != "" {
				t.Error("Python should not contribute to the build plan")
			}
		})
	})

	when("Contribute", func() {
		it("contributes python to the cache layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddPlan(buildpackplan.Plan{
				Name:     python.Dependency,
				Metadata: buildpackplan.Metadata{"build": true},
			})
			f.AddDependency(python.Dependency, stubPythonFixture)

			pythonDep, _, err := python.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = pythonDep.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(python.Dependency)

			Expect(layer).To(test.HaveLayerMetadata(true, true, false))

			Expect(filepath.Join(layer.Root, "stub-dir", "stub.txt")).To(BeARegularFile())
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONPATH", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONHOME", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONUNBUFFERED", "1"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONHASHSEED", "random"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("LANG", "en_US.UTF-8"))
		})

		it("contributes python to the launch layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddPlan(buildpackplan.Plan{
				Name:     python.Dependency,
				Metadata: buildpackplan.Metadata{"launch": true},
			})
			f.AddDependency(python.Dependency, stubPythonFixture)

			pythonContributor, _, err := python.NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = pythonContributor.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(python.Dependency)
			Expect(layer).To(test.HaveLayerMetadata(false, true, true))
			Expect(filepath.Join(layer.Root, "stub-dir", "stub.txt")).To(BeARegularFile())
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONPATH", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONHOME", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONUNBUFFERED", "1"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONHASHSEED", "random"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("LANG", "en_US.UTF-8"))
		})
	})
}
