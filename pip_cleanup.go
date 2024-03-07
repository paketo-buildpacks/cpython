package cpython

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

// This function serves as a constant for packages to be uninstalled
func pipPackagesToBeUninstalled() []string {
	return []string{"setuptools"}
}

// PipCleanup implements the PythonPipCleanup interface.
type PipCleanup struct {
	pythonProcess Executable
	logger        scribe.Emitter
}

// NewPipCleanup creates an instance of PipCleanup given a python Executable and a scribe.Emitter.
func NewPipCleanup(pythonProcess Executable, logger scribe.Emitter) PipCleanup {
	return PipCleanup{
		pythonProcess: pythonProcess,
		logger:        logger,
	}
}

func (i PipCleanup) Cleanup(packages []string, targetLayer string) error {
	env := environWithUpdatedPath(os.Environ(), "PATH", filepath.Join(targetLayer, "bin"))
	env = environWithUpdatedPath(env, "LD_LIBRARY_PATH", filepath.Join(targetLayer, "lib"))

	if len(packages) > 0 {
		// Verify pip --version works to ensure subsequent pip commands will work
		err := i.pythonProcess.Execute(pexec.Execution{
			Args:   []string{"-m", "pip", "--version"},
			Env:    env,
			Stdout: i.logger.Debug.ActionWriter,
			Stderr: i.logger.Debug.ActionWriter,
		})
		if err != nil {
			i.logger.Subprocess("pip --version failed. Run with --env BP_LOG_LEVEL=DEBUG to see more information")
			return err
		}

		// Remove packages from site-packages in the targetLayer
		for _, name := range packages {
			i.logger.Debug.Subprocess("Checking if '%s' package is installed", name)
			err := i.pythonProcess.Execute(pexec.Execution{
				Args:   []string{"-m", "pip", "show", "-q", name},
				Env:    env,
				Stdout: i.logger.Debug.ActionWriter,
				Stderr: i.logger.Debug.ActionWriter,
			})

			if err == nil {
				i.logger.Debug.Subprocess("Uninstalling '%s' package", name)
				err = i.pythonProcess.Execute(pexec.Execution{
					Args:   []string{"-m", "pip", "uninstall", "-y", name},
					Env:    env,
					Stdout: i.logger.Debug.ActionWriter,
					Stderr: i.logger.Debug.ActionWriter,
				})
				if err != nil {
					i.logger.Subprocess("pip uninstall failed. Run with --env BP_LOG_LEVEL=DEBUG to see more information")
					return err
				}
			}
		}
	}

	return nil
}
