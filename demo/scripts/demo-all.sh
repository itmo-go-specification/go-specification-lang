#!/usr/bin/env bash
# Full demo flow: Gobra verification, overlay generation, optional run.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RUN_AFTER=false
INCLUDE_FAILURE=false

usage() {
	cat <<EOF
Usage: $(basename "$0") [--run] [--include-failure]

  1. Verify demo specs with gpma:latest (gobra plugin)
  2. Generate dynamic_checks overlay
  3. Optionally run demo with overlay (--run)

Options:
  --run               Run: go run --overlay ... --section all
  --include-failure   Include failure/ files in Gobra verification
  -h, --help          Show this help
EOF
}

while [[ $# -gt 0 ]]; do
	case "$1" in
	--run)
		RUN_AFTER=true
		shift
		;;
	--include-failure)
		INCLUDE_FAILURE=true
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "unknown option: $1" >&2
		usage >&2
		exit 1
		;;
	esac
done

args=()
if [[ "$INCLUDE_FAILURE" == "true" ]]; then
	args+=(--include-failure)
fi

echo "=== Step 1: Gobra verification (gpma:latest) ==="
"${SCRIPT_DIR}/verify-gobra.sh" "${args[@]}"

echo
echo "=== Step 2: dynamic_checks overlay (gpma:latest) ==="
"${SCRIPT_DIR}/overlay-docker.sh"

if [[ "$RUN_AFTER" == "true" ]]; then
	echo
	echo "=== Step 3: Run demo with overlay ==="
	cd "${SCRIPT_DIR}/.."
	go run --overlay ../overlays/gpma_overlay.json . --section all
fi
