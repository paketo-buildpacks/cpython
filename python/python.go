package python

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const Dependency = "python"

type Contributor struct {
	buildContribution  bool
	launchContribution bool
	layer              layers.DependencyLayer
}

func NewContributor(context build.Build) (Contributor, bool, error) {
	plan, wantDependency, err := context.Plans.GetShallowMerged(Dependency)
	if err != nil || !wantDependency {
		return Contributor{}, false, err
	}

	dep, err := context.Buildpack.RuntimeDependency(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	contributor := Contributor{layer: context.Layers.DependencyLayer(dep)}
	contributor.buildContribution, _ = plan.Metadata["build"].(bool)
	contributor.launchContribution, _ = plan.Metadata["launch"].(bool)

	return contributor, true, nil
}

func (n Contributor) Contribute() error {
	return n.layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if err := helper.ExtractTarGz(artifact, layer.Root, 0); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("PYTHONPATH", layer.Root); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("PYTHONHOME", layer.Root); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("PYTHONUNBUFFERED", "1"); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("PYTHONHASHSEED", "random"); err != nil {
			return err
		}

		if err := layer.OverrideSharedEnv("LANG", "en_US.UTF-8"); err != nil {
			return err
		}

		return nil
	}, n.flags()...)
}

func (n Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if n.buildContribution {
		flags = append(flags, layers.Build)
	}

	if n.launchContribution {
		flags = append(flags, layers.Launch)
	}

	return flags
}
