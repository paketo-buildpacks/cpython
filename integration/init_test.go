package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var (
	builder       occam.Builder
	buildpackInfo struct {
		Buildpack struct {
			ID   string
			Name string
		}
		Metadata struct {
			Dependencies []struct {
				Version string
			}
		}
	}
	defaultVersion string
	settings       struct {
		Buildpacks struct {
			Cpython struct {
				Online  string
				Offline string
			}
			BuildPlan struct {
				Online string
			}
		}

		Config struct {
			BuildPlan string   `json:"build-plan"`
			Builders  []string `json:"builders"`
		}
	}
)

func TestIntegration(t *testing.T) {
	// Do not truncate Gomega matcher output
	// The buildpack output text can be large and we often want to see all of it.
	format.MaxLength = 0
	SetDefaultEventuallyTimeout(30 * time.Second)

	Expect := NewWithT(t).Expect

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&buildpackInfo)
	Expect(err).NotTo(HaveOccurred())

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()

	settings.Buildpacks.Cpython.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Cpython.Offline, err = buildpackStore.Get.
		WithVersion("1.2.3").
		WithOfflineDependencies().
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		Execute(settings.Config.BuildPlan)
	Expect(err).NotTo(HaveOccurred())

	pack := occam.NewPack().WithVerbose()
	builder, err = pack.Builder.Inspect.Execute()
	Expect(err).NotTo(HaveOccurred())

	transport := cargo.NewTransport()
	service := postal.NewService(transport)
	defaultDependency, err := service.Resolve("../buildpack.toml", "python", "default", builder.LocalInfo.Stack.ID)
	Expect(err).NotTo(HaveOccurred())
	defaultVersion = defaultDependency.Version

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	if strings.Contains(builder.LocalInfo.Stack.ID, "jammy") || strings.Contains(builder.LocalInfo.Stack.ID, "bionic") {
		suite("Offline", testOffline)
	}
	suite("LayerReuse", testLayerReuse)
	suite.Run(t)
}
