# Paketo Buildpack for CPython Cloud Native
The CPython Buildpack provides CPython (reference implementation of Python) 3.
The buildpack installs CPython onto the `$PATH` which makes it available for
subsequent buildpacks and in the final running container. It also sets the
`$PYTHONPATH` environment variable for subsequent buildpacks and launch-time
processes.

The buildpack is published for consumption at `paketobuildpacks/cpython`.

## Behavior

This buildpack always participates.

### Build time
This buildpack performs the following actions at build time:
* Contributes `cpython` to a layer - either via a precompiled dependency or by
  compiling from source.
  * Python is precompiled for recognized, supported stacks (currently the Ubuntu-based stacks)
  * Python is compiled from source during the build phase for unrecognized stacks.
  * [dependency/](dependency/README.md) contains more information how cpython is built.
* Sets the `PYTHONPYCACHEPREFIX` environment variable to the `/tmp` directory
  for this and any subsequent buildpacks.
  * This effectively disables the `__pycache__` directories, in turn enabling
    reproducible builds.
  * It can be overridden via standard means (e.g. `pack build --env
    "PYTHONPYCACHEPREFIX=/some/other/dir"`) if there is a need to keep the
    `__pycache__` directories in place of reproducible builds.
* Sets the `PYTHONPATH` to the `cpython` layer path for this and any subsequent
  buildpacks.
* Adds the newly installed `cpython` bin dir location to the `PATH`.

### Launch time
This buildpack does the following at launch time:

* Sets a default value for the `PYTHONPATH` environment variable.

## Configuration

### `BP_CPYTHON_VERSION`
The `BP_CPYTHON_VERSION` variable allows you to specify the version of CPython
that is installed. (Available versions can be found in the
[buildpack.toml](./buildpack.toml).)

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

### `BP_CPYTHON_CONFIGURE_FLAGS`
The `BP_CPYTHON_CONFIGURE_FLAGS` variable allows you to specify configure flags
when python is installed from source. This is only applicable when using custom
stacks. Paketo stacks such as `io.buildpacks.stacks.bionic` install pre-built binaries.

* The format is space-separated strings, and they are passed directly to the
  `cpython` `./configure` process , e.g. `--foo --bar=baz`.
* See [python documentation](https://docs.python.org/3/using/configure.html) for supported flags.
* Default flags if not specified: `--with-ensurepip`
* Note that default flags are overridden if you specify this environment variable,
which means you almost certainly want to include the defaults along with any custom flags.
  - e.g. `--with-ensurepip --foo --bar=baz`

### `BP_LOG_LEVEL`
The `BP_LOG_LEVEL` flag controls the level of verbosity of the buildpack output logs.
For more detail, set it to `debug`.

For example, when compiling from source, setting `BP_LOG_LEVEL=debug` shows the
commands and outputs run to build python.

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

Pre-compiled distributions of Python are provided for the Paketo stacks (i.e.
`io.buildpacks.stack.jammy` and `io.buildpacks.stacks.bionic`).

Source distributions of Python are provided for all other linux stacks.
