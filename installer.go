package cpython

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go

// CPythonInstaller implements the PythonInstaller interface.
type CPythonInstaller struct {
	configureProcess Executable
	makeProcess      Executable
	logger           scribe.Emitter
}

// NewCPythonInstaller creates an instance of the CPythonInstaller given a scribe.Emitter.
func NewCPythonInstaller(
	configureProcess Executable,
	makeProcess Executable,
	logger scribe.Emitter,
) CPythonInstaller {
	return CPythonInstaller{
		configureProcess: configureProcess,
		makeProcess:      makeProcess,
		logger:           logger,
	}
}

// Executable defines the interface for invoking an executable.
type Executable interface {
	Execute(pexec.Execution) error
}

// Installs python from source code located in the given sourcePath into the layer path designated by layerPath.
func (i CPythonInstaller) Install(
	sourcePath string,
	workingDir string,
	entry packit.BuildpackPlanEntry,
	dependency postal.Dependency,
	layerPath string,
) error {
	flags, _ := entry.Metadata["configure-flags"].(string)

	if flags == "" {
		flags = "--with-ensurepip"
		i.logger.Debug.Subprocess("Using default configure flags: %v\n", flags)
	}

	whiteSpace := regexp.MustCompile(`\s+`)
	configureFlags := whiteSpace.Split(flags, -1)
	configureFlags = append(configureFlags, "--prefix="+layerPath)

	if err := os.Chdir(sourcePath); err != nil {
		return err
	}

	i.logger.Debug.Subprocess("Running 'configure %s'", strings.Join(configureFlags, " "))
	err := i.configureProcess.Execute(pexec.Execution{
		Args:   configureFlags,
		Env:    environWithUpdatedPath(os.Environ(), "PATH", sourcePath),
		Stdout: i.logger.Debug.ActionWriter,
		Stderr: i.logger.Debug.ActionWriter,
	})
	if err != nil {
		i.logger.Subprocess("configure failed. Run with --env BP_LOG_LEVEL=DEBUG to see more information")
		return err
	}

	makeFlags := []string{"-j", fmt.Sprint(runtime.NumCPU()), `LDFLAGS="-Wl,--strip-all"`}
	i.logger.Debug.Subprocess("Running 'make %s'", strings.Join(makeFlags, " "))
	err = i.makeProcess.Execute(pexec.Execution{
		Args:   makeFlags,
		Env:    environWithUpdatedPath(os.Environ(), "PATH", sourcePath),
		Stdout: i.logger.Debug.ActionWriter,
		Stderr: i.logger.Debug.ActionWriter,
	})
	if err != nil {
		i.logger.Subprocess("make failed. Run with --env BP_LOG_LEVEL=DEBUG to see more information")
		return err
	}

	makeInstallFlags := []string{"altinstall"}
	i.logger.Debug.Subprocess("Running 'make %s'", strings.Join(makeInstallFlags, " "))
	err = i.makeProcess.Execute(pexec.Execution{
		Args:   makeInstallFlags,
		Env:    environWithUpdatedPath(os.Environ(), "PATH", sourcePath),
		Stdout: i.logger.Debug.ActionWriter,
		Stderr: i.logger.Debug.ActionWriter,
	})
	if err != nil {
		i.logger.Subprocess("make install failed. Run with --env BP_LOG_LEVEL=DEBUG to see more information")
		return err
	}

	versionList := strings.Split(dependency.Version, ".")
	major := versionList[0]
	majorMinor := strings.Join(versionList[:len(versionList)-1], ".")

	if err = os.Chdir(filepath.Join(layerPath, "bin")); err != nil {
		return err
	}

	for _, name := range []string{"python", "pip"} {
		i.logger.Debug.Action("Writing symlink bin/%s", name+major)
		if err = os.Symlink(name+majorMinor, name+major); err != nil {
			return err
		}
	}

	if err = os.Chdir(workingDir); err != nil {
		return err
	}

	return nil
}

// Returns environment variables with customPath inserted at the beginning of given environment variable
func environWithUpdatedPath(environ []string, variableName, customPath string) []string {
	var env []string = nil
	varNameWithEqual := variableName + "="
	environmentVariableExists := false

	for _, v := range environ {
		if strings.HasPrefix(v, varNameWithEqual) {
			env = append(env, varNameWithEqual+customPath+":"+strings.TrimPrefix(v, varNameWithEqual))
			environmentVariableExists = true
		} else {
			env = append(env, v)
		}
	}

	if !environmentVariableExists {
		env = append(env, varNameWithEqual+customPath)
	}

	return env
}
