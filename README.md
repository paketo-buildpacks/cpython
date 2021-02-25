# Paketo CPython Cloud Native Buildpack

The CPython Buildpack provides CPython (reference implementation of Python) 3.
The buildpack installs CPython onto the `$PATH` which makes it available for
subsequent buildpacks and in the final running container. It also sets the
`$PYTHONPATH` environment variable.

## Integration

The CPython Buildpack provides `cpython` as a dependency. Downstream
buildpacks, can require the cpython dependency by generating [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]
  # The name of the CPython dependency is "cpython". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "cpython"
  # The version of the CPython dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "3.*", "3.8.*", or even
  # "3.8.2".
  version = "3.8.2"
  # The CPython buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the CPython
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to use CPython
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that cpython is
    # available to the running application. If you are writing an application
    # that needs to run cpython at runtime, this flag should be set to true.
    launch = true
```

## Usage

To package this buildpack for consumption:

```
$ ./scripts/package.sh --version <version-number>
```

This will create a `buildpackage.cnb` file under the `build` directory which you
can use to build your app as follows:
`pack build <app-name> -p <path-to-app> -b build/buildpackage.cnb -b <other-buildpacks..>`

## `buildpack.yml` Configurations

In order to specify a particular version of python you can
provide an optional `buildpack.yml` in the root of the application directory.

```yaml
cpython:
  # this allows you to specify a version constraint for the python depdendency
  # any valid semver constaints (e.g. 3.*) are also acceptable
  version: ~3
```



