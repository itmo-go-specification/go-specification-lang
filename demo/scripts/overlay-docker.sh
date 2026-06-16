#!/usr/bin/env bash
# Generate dynamic_checks overlay via gpma:latest Docker image.
set -euo pipefail

IMAGE="${GPMA_IMAGE:-gpma:latest}"
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
WORK_MOUNT="${GPMA_WORK_MOUNT:-/work}"
OUTPUT_DIR="${1:-${REPO_ROOT}/overlays}"

mkdir -p "$OUTPUT_DIR"

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
	echo "Docker image '$IMAGE' not found. Build it from the repo root:" >&2
	echo "  just docker-gpma" >&2
	exit 1
fi

container_out="${WORK_MOUNT}/overlays"
if [[ "$OUTPUT_DIR" != "${REPO_ROOT}/overlays" ]]; then
	echo "Custom output directories are not supported with Docker mount." >&2
	echo "Overlay will be written to ${REPO_ROOT}/overlays/" >&2
	container_out="${WORK_MOUNT}/overlays"
fi

docker run --rm -v "${REPO_ROOT}:${WORK_MOUNT}" -w "${WORK_MOUNT}" "$IMAGE" \
	project -p dynamic_checks "${WORK_MOUNT}/demo/" "${container_out}/"

echo "Overlay written to ${REPO_ROOT}/overlays/"
echo "Run with: cd demo && go run --overlay ../overlays/gpma_overlay.json . --section all"
