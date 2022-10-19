#!/bin/bash

set -euo pipefail
shopt -s inherit_errexit

parent_dir="$(cd "$(dirname "$0")" && pwd)"

extract_tarball() {
  rm -rf cpython
  mkdir cpython
  tar --extract --file "${1}" \
    --directory cpython
}

set_ld_library_path() {
  export LD_LIBRARY_PATH="$PWD/cpython/lib:${LD_LIBRARY_PATH:-}"
}

check_version() {
  expected_version="${1}"
  actual_version="$(./cpython/bin/python3 --version | cut -d' ' -f2)"
  if [[ "${actual_version}" != "${expected_version}" ]]; then
    echo "Version ${actual_version} does not match expected version ${expected_version}"
    exit 1
  fi
}

check_server() {
  set +e

  ./cpython/bin/python3 "${parent_dir}/fixtures/server.py" 8080 &
  server_pid=$!

  succeeded=0
  for _ in {1..5}; do
    response="$(curl -s http://localhost:8080)"
    if [[ $response == *"Hello world!"* ]]; then
      succeeded=1
      break
    fi
    sleep 1
  done

  kill "${server_pid}"

  if [[ ${succeeded} -eq 0 ]]; then
    echo "Failed to curl server"
    exit 1
  fi

  set -e
}

main() {
  local tarballPath expectedVersion
  tarballPath=""
  expectedVersion=""

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --tarballPath)
        tarballPath="${2}"
        shift 2
        ;;

      --expectedVersion)
        expectedVersion="${2}"
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

  if [[ "${tarballPath}" == "" ]]; then
    echo "--tarballPath is required"
    exit 1
  fi

  if [[ "${expectedVersion}" == "" ]]; then
    echo "--expectedVersion is required"
    exit 1
  fi

  echo "tarballPath=${tarballPath}"
  echo "expectedVersion=${expectedVersion}"

  extract_tarball "${tarballPath}"
  set_ld_library_path
  check_version "${expectedVersion}"
  check_server

  echo "All tests passed!"
}

main "$@"
