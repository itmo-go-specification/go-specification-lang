#!/usr/bin/env bash
# Batch Gobra verification via gpma:latest Docker image.
set -euo pipefail

IMAGE="${GPMA_IMAGE:-gpma:latest}"
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
WORK_MOUNT="${GPMA_WORK_MOUNT:-/work}"
INCLUDE_FAILURE=false

usage() {
	cat <<EOF
Usage: $(basename "$0") [--include-failure]

Verify demo Go files with: docker run ... gpma file --plugin gobra

Options:
  --include-failure   Also verify failure/ files (expected to fail verification)
  -h, --help          Show this help

Environment:
  GPMA_IMAGE        Docker image (default: gpma:latest)
  GPMA_WORK_MOUNT   Mount path inside container (default: /work)
EOF
}

while [[ $# -gt 0 ]]; do
	case "$1" in
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

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
	echo "Docker image '$IMAGE' not found. Build it from the repo root:" >&2
	echo "  just docker-gpma" >&2
	exit 1
fi

verify_file() {
	local rel="$1"
	local expect_fail="${2:-false}"
	local path="${WORK_MOUNT}/${rel}"

	echo "==> $rel"
	set +e
	output="$(docker run --rm -v "${REPO_ROOT}:${WORK_MOUNT}" -w "${WORK_MOUNT}" "$IMAGE" \
		file --plugin gobra "$path" 2>&1)"
	status=$?
	set -e

	if [[ "$expect_fail" == "true" ]]; then
		if echo "$output" | grep -qiE 'verification failed|postcondition|precondition|might not hold|might fail|\[gobra\].*error:'; then
			echo "    OK (expected verification failure)"
			return 0
		fi
		echo "$output"
		echo "    FAIL: expected verification failure for $rel" >&2
		return 1
	fi

	if [[ $status -eq 0 ]] && ! echo "$output" | grep -qi 'plugin gobra failed'; then
		echo "    OK"
		return 0
	fi

	echo "$output"
	echo "    FAIL: verification failed for $rel" >&2
	return 1
}

failures=0

while IFS= read -r -d '' file; do
	rel="${file#${REPO_ROOT}/}"
	if [[ "$rel" == demo/main.go ]]; then
		continue
	fi
	if [[ "$rel" == *"/failure/"* ]]; then
		continue
	fi
	verify_file "$rel" false || failures=$((failures + 1))
done < <(find "${REPO_ROOT}/demo" -name '*.go' -print0 | sort -z)

if [[ "$INCLUDE_FAILURE" == "true" ]]; then
	while IFS= read -r -d '' file; do
		rel="${file#${REPO_ROOT}/}"
		verify_file "$rel" true || failures=$((failures + 1))
	done < <(find "${REPO_ROOT}/demo" -path '*/failure/*.go' -print0 | sort -z)
fi

if [[ $failures -gt 0 ]]; then
	echo "${failures} file(s) failed" >&2
	exit 1
fi

echo "All verifications passed."
