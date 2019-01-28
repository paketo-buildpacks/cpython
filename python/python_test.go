package python

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/test"
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
	when("NewContributor", func() {
		var stubPythonFixture = filepath.Join("testdata", "stub-python.tar.gz")

		it("returns true if python is in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{})
			f.AddDependency(Dependency, stubPythonFixture)

			_, willContribute, err := NewContributor(f.Build)
			if err != nil {
				t.Error("Could not create contributor")
			}

			if diff := cmp.Diff(willContribute, true); diff != "" {
				t.Error("Python should contribute to the build plan")
			}

		})

		it("returns false if python is not in the build plan", func() {
			f := test.NewBuildFactory(t)

			_, willContribute, err := NewContributor(f.Build)
			if err != nil {
				t.Error("Could not create contributor")
			}

			if diff := cmp.Diff(willContribute, false); diff != "" {
				t.Error("Python should not contribute to the build plan")
			}
		})

		it("contributes python to the cache layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"build": true},
			})
			f.AddDependency(Dependency, stubPythonFixture)

			pythonDep, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = pythonDep.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(Dependency)

			Expect(layer).To(test.HaveLayerMetadata(true, true, false))

			Expect(filepath.Join(layer.Root, "stub-dir", "stub.txt")).To(BeARegularFile())
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONPATH", layer.Root)) //s.Stager.DepDir()
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONHOME", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONUNBUFFERED", "1"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONHASHSEED", "random"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("LANG", "en_US.UTF-8"))

		})

		it("contributes python to the launch layer when included in the build plan", func() {
			f := test.NewBuildFactory(t)
			f.AddBuildPlan(Dependency, buildplan.Dependency{
				Metadata: buildplan.Metadata{"launch": true},
			})
			f.AddDependency(Dependency, stubPythonFixture)

			pythonContributor, _, err := NewContributor(f.Build)
			Expect(err).NotTo(HaveOccurred())

			err = pythonContributor.Contribute()
			Expect(err).NotTo(HaveOccurred())

			layer := f.Build.Layers.Layer(Dependency)
			Expect(layer).To(test.HaveLayerMetadata(false, true, true))
			Expect(filepath.Join(layer.Root, "stub-dir", "stub.txt")).To(BeARegularFile())
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONPATH", layer.Root)) //s.Stager.DepDir()
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONHOME", layer.Root))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONUNBUFFERED", "1"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("PYTHONHASHSEED", "random"))
			Expect(layer).To(test.HaveOverrideSharedEnvironment("LANG", "en_US.UTF-8"))
		})
	})
}
