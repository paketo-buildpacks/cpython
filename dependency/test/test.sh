#!/usr/bin/env bash

set -euo pipefail

if [[ "$(uname)" != "Darwin" ]]; then
  shopt -s inherit_errexit
fi

parent_dir="$(cd "$(dirname "$0")" && pwd)"

main() {
  local tarball_path expectedVersion
  tarball_path=""
  expectedVersion=""
  expectedOS=""
  expectedArch=""

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --tarballPath)
        tarball_path="${2}"
        shift 2
        ;;

      --expectedVersion)
        expectedVersion="${2}"
        shift 2
        ;;

      --expectedOS)
        expectedOS="${2}"
        shift 2
        ;;

      --expectedArch)
        expectedArch="${2}"
        shift 2
        ;;

      "")
        shift
        ;;

      *)
        echo "unknown argument \"${1}\""
        exit 1
    esac
  done

  if [[ "${tarball_path}" == "" ]]; then
    echo "--tarballPath is required"
    exit 1
  fi

  if [[ "${expectedVersion}" == "" ]]; then
    echo "--expectedVersion is required"
    exit 1
  fi

  echo "Outside image: tarball_path=${tarball_path}"
  echo "Outside image: expectedVersion=${expectedVersion}"
  echo "Outside image: expectedOS=${expectedOS}"
  echo "Outside image: expectedArch=${expectedArch}"

  # Expects tarball_path filename such as: python_3.10.18_linux_amd64_bionic_ed1b62b3.tgz
  tarballVersion=$(echo $(basename "${tarball_path}") | cut -d_ -f2) # ie: 3.10.18
  os=$(echo $(basename "${tarball_path}") | cut -d_ -f3 | tr '_' '/') # ie: linux
  arch=$(echo $(basename "${tarball_path}") | cut -d_ -f4 | tr '_' '/') # ie: amd64
  target=$(echo $(basename "${tarball_path}") | cut -d_ -f5) # ie: bionic

  if [[ "${tarballVersion}" != "${expectedVersion}" ]]; then
    echo "Tarball version (${tarballVersion}) does not match expected version (${expectedVersion})"
    exit 1
  fi

  if [[ "${expectedOS}" != "" && "${expectedOS}" != "${os}" ]]; then
    echo "Tarball os (${os}) does not match expectedOS (${expectedOS})"
    exit 1
  fi

  if [[ "${expectedArch}" != "" && "${expectedArch}" != "${arch}" ]]; then
    echo "Tarball arch (${arch}) does not match expectedArch (${expectedArch})"
    exit 1
  fi

  # When --expectedOS and --expectedArch are provided, the --platform arg is passed to docker build and run commands.
  # This assumes the runner has qemu and buildkit set up, and that the docker daemon and cli experimental features are enabled.
  docker_platform_arg=""
  if [[ "${expectedOS}" != "" && "${expectedArch}" != "" ]]; then
    docker_platform_arg="--platform ${os}/${arch}"
    echo "Outside image: docker commands will be called with ${docker_platform_arg}"
  fi

  if [[ -f "$target.Dockerfile" ]]; then
    echo "Building image with dockerfile ${target}.Dockerfile..."
    docker build \
      --quiet \
      --tag test \
      --file "${target}.Dockerfile" \
      ${docker_platform_arg} \
      .

    echo "Running ${target} test..."
    docker run \
      --rm \
      --volume "$(dirname -- "${tarball_path}"):/tarball_path" \
      ${docker_platform_arg} \
      test \
      --tarballPath "/tarball_path/$(basename "${tarball_path}")" \
      --expectedVersion "${expectedVersion}"
  else
    echo "No dockerfile found for ${target}, expected ${target}.Dockerfile to exist"
    exit 1
  fi
}

main "$@"
