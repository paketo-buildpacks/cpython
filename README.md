# Paketo CPython Cloud Native Buildpack
The CPython Buildpack provides CPython (reference implementation of Python) 3.
The buildpack installs CPython onto the `$PATH` which makes it available for
subsequent buildpacks and in the final running container. It also sets the
`$PYTHONPATH` environment variable.

The buildpack is published for consumption at `gcr.io/paketo-community/cpython` and
`paketocommunity/cpython`.

## Behavior
This buildpack always participates.

The buildpack will do the following:
* At build time:
  - Contributes `cpython` to a layer
  - Sets the `PYTHONPATH` to the `cpython` layer path.
  - Adds the newly installed `cpython` bin dir location to `PATH`
* At run time:
  - Does nothing

## Configuration

### `BP_CPYTHON_VERSION`
The `BP_CPYTHON_VERSION` variable allows you to specify the version of CPython
that is installed. (Available versions can be found in the
[buildpack.toml](./buildpack.toml).)

Specifying the CPython version through `buildpack.yml` configuration
is not supported from CPython Buildpack v1.0.0 onwards.

#### `pack build` flag
```shell
pack build my-app --env BP_CPYTHON_VERSION=3.6.*
```

#### In a [`project.toml`](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md)
```toml
[build]
  [[build.env]]
    name = 'BP_CPYTHON_VERSION'
    value = '3.6.*' # any valid semver constraints (e.g. 3.6.7, 3.*) are acceptable
```

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

To run the unit and integration tests for this buildpack:
```
$ ./scripts/unit.sh && ./scripts/integration.sh
```
