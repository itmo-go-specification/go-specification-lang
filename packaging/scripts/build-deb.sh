#!/usr/bin/env bash
# Assemble and build the gpma .deb package.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST="${ROOT}/packaging/dist"
STAGING="${ROOT}/packaging/staging"
FILE_ROOT="${ROOT}/packaging/files"

# shellcheck source=packaging/versions.env
source "${ROOT}/packaging/versions.env"

"${SCRIPT_DIR}/fetch-deps.sh"
"${SCRIPT_DIR}/build-gpma.sh"

mkdir -p "${DIST}/deb-root/usr/bin" \
         "${DIST}/deb-root/usr/lib/gpma/gobra" \
         "${DIST}/deb-root/usr/lib/gpma/bin" \
         "${DIST}/deb-root/usr/lib/gpma/lib" \
         "${DIST}/deb-root/etc/gpma" \
         "${DIST}/deb-root/usr/share/doc/gpma" \
         "${DIST}/deb-root/usr/share/gpma/examples"

install -m 0755 "${DIST}/gpma" "${DIST}/deb-root/usr/bin/gpma"
install -m 0644 "${STAGING}/gobra/gobra.jar" "${DIST}/deb-root/usr/lib/gpma/gobra/gobra.jar"
install -m 0755 "${STAGING}/bin/z3" "${DIST}/deb-root/usr/lib/gpma/bin/z3"
install -m 0755 "${STAGING}/bin/Boogie" "${DIST}/deb-root/usr/lib/gpma/bin/Boogie"
install -m 0644 "${STAGING}/lib/libssl.so.1.1" "${DIST}/deb-root/usr/lib/gpma/lib/libssl.so.1.1"
install -m 0644 "${STAGING}/lib/libcrypto.so.1.1" "${DIST}/deb-root/usr/lib/gpma/lib/libcrypto.so.1.1"
install -m 0755 "${FILE_ROOT}/gobra-wrapper.sh" "${DIST}/deb-root/usr/lib/gpma/gobra/wrapper.sh"
install -m 0644 "${FILE_ROOT}/gobra.env" "${DIST}/deb-root/etc/gpma/gobra.env"
install -m 0644 "${ROOT}/LICENSE" "${DIST}/deb-root/usr/share/doc/gpma/LICENSE"
install -m 0644 "${FILE_ROOT}/copyright" "${DIST}/deb-root/usr/share/doc/gpma/copyright"
install -m 0644 "${FILE_ROOT}/smoke.gobra" "${DIST}/deb-root/usr/share/gpma/examples/smoke.gobra"

ln -sf /usr/lib/gpma/gobra/wrapper.sh "${DIST}/deb-root/usr/bin/gobra"

if ! command -v nfpm >/dev/null 2>&1; then
  echo "Installing nfpm..."
  go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.41.2
  export PATH="${PATH}:$(go env GOPATH)/bin"
fi

NFPM_CONFIG="${DIST}/nfpm.yaml"
sed "s/@VERSION@/${GPMA_VERSION}/g" "${ROOT}/packaging/nfpm.yaml" > "${NFPM_CONFIG}"

mkdir -p "${DIST}"
nfpm pkg --config "${NFPM_CONFIG}" --packager deb --target "${DIST}/"

echo "Package written to ${DIST}/"
