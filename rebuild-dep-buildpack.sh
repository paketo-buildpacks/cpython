#!/usr/bin/env bash

set -eu
set -o pipefail
shopt -s inherit_errexit

mkdir -p /tmp/jammy-dep

# docker build --tag compilation-jammy --file ./dependency/actions/compile/jammy.Dockerfile ./dependency/actions/compile
# docker run --volume /tmp/jammy-dep:/tmp/compilation compilation-jammy --outputDir /tmp/compilation --target jammy --version 3.11.0
# mv /tmp/jammy-dep/python_3.11.0_linux_x64_jammy_*.tgz /tmp/jammy-dep/python_3.11.0_linux_x64_jammy.tgz
# mv /tmp/jammy-dep/python_3.11.0_linux_x64_jammy_*.tgz.checksum /tmp/jammy-dep/python_3.11.0_linux_x64_jammy.tgz.checksum
checksum=$(cat /tmp/jammy-dep/*.checksum)
sed -i '' "19s/.*/    checksum = \"${checksum}\"/" buildpack.toml
./scripts/package.sh --version 9.8.7

pack build \
    --clear-cache \
    --verbose \
    --builder paketobuildpacks/builder-jammy-buildpackless-base:latest \
    -b ~/workspace/paketo-buildpacks/cpython/build/buildpack.tgz \
    -b index.docker.io/paketocommunity/build-plan \
    --env BP_CPYTHON_VERSION='3.11' \
    --path ~/workspace/paketo-buildpacks/cpython/integration/testdata/default_app \
    cpython-test


pack build \
    --clear-cache \
    --verbose \
    --builder paketobuildpacks/builder-jammy-buildpackless-base:latest \
    -b ~/workspace/paketo-buildpacks/cpython/build/buildpack.tgz \
    -b index.docker.io/paketobuildpacks/pip \
    -b index.docker.io/paketobuildpacks/pip-install \
    -b index.docker.io/paketocommunity/build-plan \
    --env BP_CPYTHON_VERSION='3.11' \
    --path ~/workspace/paketo-buildpacks/pip-install/integration/testdata/default_app \
    pip-install-test

