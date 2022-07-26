#!/usr/bin/env bash

set -eu
set -o pipefail

WORKING_DIR=$(mktemp -d)
DEST_DIR=$(mktemp -d)
OUTPUT_DIR=$(mktemp -d)

function scream {
  echo "--- --- --- ---"
  echo "--- --- --- ---"
  echo "--- --- --- ---"
  echo "${0}"
  echo "--- --- --- ---"
  echo "--- --- --- ---"
  echo "--- --- --- ---"
}

scream "Running apt-get"

sudo apt-get update
sudo apt-get -y install libdb-dev
sudo apt-get -y install libgdbm-dev
sudo apt-get -y install tk8.6-dev
sudo apt-get -y --force-yes -d install --reinstall libtcl8.6 libtk8.6 libxss1

dpkg -x /var/cache/apt/archives/libtcl8.6_*.deb "${DEST_DIR}"
dpkg -x /var/cache/apt/archives/libtk8.6_*.deb "${DEST_DIR}"
dpkg -x /var/cache/apt/archives/libxss1_*.deb "${DEST_DIR}"

pushd "${WORKING_DIR}"

  scream "Downloading upstream tarball"

  curl "${UPSTREAM_TARBALL}" \
    --silent \
    --output upstream.tgz

  UPSTREAM_SHA256=$(sha256sum upstream.tgz)
  UPSTREAM_SHA256=${UPSTREAM_SHA256:0:64}
  OUTPUT_TARBALL_NAME="python-${VERSION}-linux_x64_jammy_${UPSTREAM_SHA256:0:8}.tgz"

  tar --extract \
    --file upstream.tgz

  pushd "Python-${VERSION}"
    scream "Running Python's ./configure script"

    ./configure \
      --enable-shared \
      --with-ensurepip=yes \
      --with-dbmliborder=bdb:gdbm \
      --with-tcltk-includes="-I/usr/include/tcl8.6" \
      --with-tcltk-libs="-L/usr/lib/x86_64-linux-gnu -ltcl8.6 -L/usr/lib/x86_64-linux-gnu -ltk8.6" \
      --prefix="${DEST_DIR}" \
      --enable-unicode=ucs4

    scream "Running make and make install"

    make
    make install
  popd
popd

pushd "${DEST_DIR}"
    scream "Building tarball ${OUTPUT_TARBALL_NAME}"

    tar --create \
      --gzip \
      --verbose \
      --hard-dereference \
      --file "${OUTPUT_DIR}/${OUTPUT_TARBALL_NAME}" \
      .
popd

echo "::set-output name=upstream-sha256::${UPSTREAM_SHA256}"
echo "::set-output name=tarball-name::${OUTPUT_TARBALL_NAME}"
echo "::set-output name=tarball-path::${OUTPUT_DIR}/${OUTPUT_TARBALL_NAME}"

SHA256=$(sha256sum "${OUTPUT_DIR}/${OUTPUT_TARBALL_NAME}")
SHA256="${SHA256:0:64}"
echo "::set-output name=sha256::${SHA256}"
