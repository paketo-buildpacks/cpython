#!/usr/bin/env bash

set -eu
set -o pipefail
shopt -s inherit_errexit

DEST_DIR=$(mktemp -d)

dpkg -x /var/cache/apt/archives/libtcl8.6_*.deb "${DEST_DIR}"
dpkg -x /var/cache/apt/archives/libtk8.6_*.deb "${DEST_DIR}"
dpkg -x /var/cache/apt/archives/libxss1_*.deb "${DEST_DIR}"

function main() {
  local version output_dir target upstream_tarball working_dir
  version=""
  output_dir=""
  target=""
  # These were the original values before os and arch args were added
  os="linux"
  arch="x64"
  upstream_tarball=""
  working_dir=$(mktemp -d)

  while [ "${#}" != 0 ]; do
    case "${1}" in
      --version)
        version="${2}"
        shift 2
        ;;

      --outputDir)
        output_dir="${2}"
        shift 2
        ;;

      --target)
        target="${2}"
        shift 2
        ;;

      --os)
        os="${2}"
        shift 2
        ;;

      --arch)
        arch="${2}"
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

  if [[ "${version}" == "" ]]; then
    echo "--version is required"
    exit 1
  fi

  if [[ "${output_dir}" == "" ]]; then
    echo "--outputDir is required"
    exit 1
  fi

  if [[ "${target}" == "" ]]; then
    echo "--target is required"
    exit 1
  fi

  echo "version=${version}"
  echo "output_dir=${output_dir}"
  echo "target=${target}"
  echo "os=${os}"
  echo "arch=${arch}"

  pushd "${working_dir}" > /dev/null
    upstream_tarball="https://www.python.org/ftp/python/${version}/Python-${version}.tgz"

    echo "Downloading upstream tarball from ${upstream_tarball}"

    curl "${upstream_tarball}" \
      --silent \
      --fail \
      --output upstream.tgz

    tar --extract \
      --file upstream.tgz

    pushd "Python-${version}" > /dev/null
      echo "Running Python's ./configure script"

      ./configure \
        --enable-shared \
        --with-ensurepip=yes \
        --with-dbmliborder=bdb:gdbm \
        --with-tcltk-includes="-I/usr/include/tcl8.6" \
        --with-tcltk-libs="-L/usr/lib/$(arch)-linux-gnu -ltcl8.6 -L/usr/lib/$(arch)-linux-gnu -ltk8.6" \
        --with-openssl="/usr/" \
        --prefix="${DEST_DIR}" \
        --enable-unicode=ucs4

      echo "Running make and make install"

      make -j$(nproc) LDFLAGS="-Wl,--strip-all"
      make install
    popd > /dev/null
  popd > /dev/null

  pushd "${DEST_DIR}" > /dev/null
    tar --create \
      --gzip \
      --verbose \
      --hard-dereference \
      --file "${output_dir}/temp.tgz" \
      .
  popd > /dev/null

  pushd "${output_dir}" > /dev/null
    local sha256
    sha256=$(sha256sum temp.tgz)
    sha256="${sha256:0:64}"

    output_tarball_name="python_${version}_${os}_${arch}_${target}_${sha256:0:8}.tgz"

    echo "Building tarball ${output_tarball_name}"

    mv temp.tgz "${output_tarball_name}"
    echo "sha256:${sha256}" > "${output_tarball_name}.checksum"
  popd > /dev/null
}

main "${@:-}"
