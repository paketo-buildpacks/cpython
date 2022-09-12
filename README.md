# Paketo Buildpack for CPython Cloud Native
The CPython Buildpack provides CPython (reference implementation of Python) 3.
The buildpack installs CPython onto the `$PATH` which makes it available for
subsequent buildpacks and in the final running container. It also sets the
`$PYTHONPATH` environment variable.

The buildpack is published for consumption at `gcr.io/paketo-buildpacks/cpython` and
`paketobuildpacks/cpython`.

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

### `BP_CPYTHON_CONFIGURE_FLAGS`
The `BP_CPYTHON_CONFIGURE_FLAGS` variable allows you to specify configure flags
when python is installed from source. This is only applicable when using custom 
stacks. Paketo stacks such as `io.buildpacks.stacks.bionic` install pre-built binaries. 

* The format is space-separated strings, and they are passed directly to the `cpython` `./configure` process , e.g. `--foo --bar=baz`.
* See [python documentation](https://docs.python.org/3/using/configure.html) for supported flags.
* Default flags if not specified: `--enable-optimizations --with-ensurepip`
* Note that default flags are overridden if you specify this environment variable,
which means you almost certainly want to include the defaults along with any custom flags.
  - e.g. `--enable-optimizations --with-ensurepip --foo --bar=baz`

### `BP_LOG_LEVEL`
When using custom stacks that install python from source setting `BP_LOG_LEVEL=DEBUG`
shows the commands and outputs run to build python.

#### `pack build` flag
```shell
pack build my-app --env BP_CPYTHON_VERSION=3.10.*
```

#### In a [`project.toml`](https://github.com/buildpacks/spec/blob/main/extensions/project-descriptor.md)
```toml
[build]
  [[build.env]]
    name = 'BP_CPYTHON_VERSION'
    value = '3.10.*' # any valid semver constraints (e.g. 3.10.2, 3.*) are acceptable
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

  # The CPython buildpack supports some non-required metadata options.
  [requires.metadata]

    # The version of the CPython dependency is not required. In the case it
    # is not specified, the buildpack will provide the default version, which can
    # be seen in the buildpack.toml file.
    # If you wish to request a specific version, the buildpack supports
    # specifying a semver constraint in the form of "3.*", "3.10.*", or even
    # "3.10.2".
    version = "3.10.2"

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

## Compatibility

This buildpack is currently only supported on linux distributions.
