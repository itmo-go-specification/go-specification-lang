#!/usr/bin/env bash
# Download Gobra JAR, Z3, and Boogie into packaging/staging/.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
STAGING="${ROOT}/packaging/staging"
CACHE="${ROOT}/packaging/cache"

# shellcheck source=packaging/versions.env
source "${ROOT}/packaging/versions.env"

mkdir -p "${STAGING}/gobra" "${STAGING}/bin" "${STAGING}/lib" "${CACHE}"

ensure_crane() {
  if command -v crane >/dev/null 2>&1; then
    return
  fi
  echo "Installing crane..."
  go install github.com/google/go-containerregistry/cmd/crane@v0.20.3
  export PATH="${PATH}:$(go env GOPATH)/bin"
}

fetch_gobra_jar() {
  local jar="${STAGING}/gobra/gobra.jar"
  if [[ -f "${jar}" ]]; then
    echo "Gobra JAR already present: ${jar}"
    return
  fi

  echo "Fetching Gobra JAR from ${GOBRA_DOCKER_IMAGE}..."

  if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    local cid
    cid="$(docker create "${GOBRA_DOCKER_IMAGE}")"
    trap 'docker rm -f "${cid}" >/dev/null 2>&1 || true' EXIT
    docker cp "${cid}:/gobra/gobra.jar" "${jar}"
    docker rm -f "${cid}" >/dev/null
    trap - EXIT
  else
    ensure_crane
    crane export "${GOBRA_DOCKER_IMAGE}" - | tar -xO gobra/gobra.jar > "${jar}"
  fi

  echo "Saved ${jar}"
}

fetch_z3() {
  local dest="${STAGING}/bin/z3"
  if [[ -f "${dest}" ]]; then
    echo "Z3 already present: ${dest}"
    return
  fi

  local zip="${CACHE}/z3.zip"
  if [[ ! -f "${zip}" ]]; then
    echo "Downloading Z3 from ${Z3_URL}..."
    curl -fsSL "${Z3_URL}" -o "${zip}"
  fi

  echo "Extracting Z3..."
  rm -rf "${CACHE}/z3"
  unzip -qo "${zip}" -d "${CACHE}/z3"
  install -m 0755 "${CACHE}/z3/${Z3_BIN_PATH}" "${dest}"
  echo "Saved ${dest}"
}

fetch_boogie() {
  local dest="${STAGING}/bin/Boogie"
  if [[ -f "${dest}" ]]; then
    echo "Boogie already present: ${dest}"
    return
  fi

  local zip="${CACHE}/boogie.zip"
  if [[ ! -f "${zip}" ]]; then
    echo "Downloading Boogie from ${BOOGIE_URL}..."
    curl -fsSL "${BOOGIE_URL}" -o "${zip}"
  fi

  echo "Extracting Boogie..."
  rm -rf "${CACHE}/boogie"
  unzip -qo "${zip}" -d "${CACHE}/boogie"
  install -m 0755 "${CACHE}/boogie/${BOOGIE_BIN_PATH}" "${dest}"
  echo "Saved ${dest}"
}

fetch_libssl() {
  local libssl="${STAGING}/lib/libssl.so.1.1"
  local libcrypto="${STAGING}/lib/libcrypto.so.1.1"
  if [[ -f "${libssl}" && -f "${libcrypto}" ]]; then
    echo "OpenSSL 1.1 libs already present in ${STAGING}/lib"
    return
  fi

  local deb="${CACHE}/libssl1.1.deb"
  if [[ ! -f "${deb}" ]]; then
    echo "Downloading libssl1.1 from ${LIBSSL_DEB_URL}..."
    curl -fsSL "${LIBSSL_DEB_URL}" -o "${deb}"
  fi

  echo "Extracting OpenSSL 1.1 compatibility libraries..."
  rm -rf "${CACHE}/libssl"
  mkdir -p "${CACHE}/libssl"
  if command -v dpkg-deb >/dev/null 2>&1; then
    dpkg-deb -x "${deb}" "${CACHE}/libssl"
  else
    local tmp
    tmp="$(mktemp -d)"
    (cd "${tmp}" && ar x "${deb}" && tar -xf data.tar.*)
    mv "${tmp}/usr" "${CACHE}/libssl/"
    rm -rf "${tmp}"
  fi
  install -m 0644 "${CACHE}/libssl/usr/lib/x86_64-linux-gnu/libssl.so.1.1" "${libssl}"
  install -m 0644 "${CACHE}/libssl/usr/lib/x86_64-linux-gnu/libcrypto.so.1.1" "${libcrypto}"
  echo "Saved ${libssl} and ${libcrypto}"
}

fetch_gobra_jar
fetch_z3
fetch_boogie
fetch_libssl

echo "Dependencies staged in ${STAGING}"
