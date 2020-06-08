# Python Runtime Cloud Native Buildpack

The Python Runtime CNB provides the Python 3 runtime.
The buildpack installs Python onto the `$PATH` which makes it available
for subsequent buildpacks and in the final running container. It also sets
the `$PYTHONPATH` environment variable.

## Integration

The Python Runtime CNB provides `python` as a dependency. Downstream buildpacks,
can require the python dependency by generating
[Build Plan TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]
  # The name of the Python dependency is "python". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "python"
  # The version of the Python dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "3.*", "3.8.*", or even
  # "3.8.2".
  version = "3.8.2"
  # The Python buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the Python
    # depdendency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to use Python
    # during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the python runtime is
    # available to the running application. If you are writing an application that needs to run
    # python at runtime, this flag should be set to true.
    launch = true
```

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## `buildpack.yml` Configurations

In order to specify a particular version of python you can
provide an optional `buildpack.yml` in the root of the application directory.

```yaml
python:
  # this allows you to specify a version constraint for the python depdendency
  # any valid semver constaints (e.g. 3.*) are also acceptable
  version: ~3
```
