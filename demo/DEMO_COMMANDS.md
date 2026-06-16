# GPMA Demo - Command Reference

All GPMA commands below use the **`gpma:latest`** Docker image from this repository.
Gobra runs **natively inside the image** — do not pass `gobra:backend:docker`.

## Prerequisites

```bash
# From repository root
just docker-gpma
# or faster rebuild:
just docker-gpma-runtime
```

Verify the image:

```bash
docker run --rm gpma:latest plugins
```

## Build Demo (on host)

```bash
cd demo && go build ./...
go run . --section all
```

## Gobra — static verification

Single file:

```bash
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/math/gcd.go
```

**Expected:** no fatal plugin errors; verification succeeds for correct contracts.

### Phase 1 – core

```bash
# Math
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/math/gcd.go
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/math/abs.go
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/math/clamp.go

# Search & sort
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/search/binary.go
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/sort/insertion.go

# Order pricing
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/order/pricing.go
```

### Phase 2 – extended

```bash
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/math/factorial.go
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/math/fibonacci.go
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/strings/hamming.go
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/dynamic/maxsubarray.go
```

### Failure examples (should fail verification)

```bash
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/failure/binary_search_wrong.go
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/order/failure/discount_wrong.go
```

### Batch verification (script)

From repository root:

```bash
./demo/scripts/verify-gobra.sh
./demo/scripts/verify-gobra.sh --include-failure
```

## dynamic_checks — runtime contract overlay

### Generate overlay

```bash
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  project -p dynamic_checks /work/demo/ /work/overlays/
```

Or use the helper script:

```bash
./demo/scripts/overlay-docker.sh
```

### Run with overlay (on host, requires Go toolchain)

```bash
cd demo
go run --overlay ../overlays/gpma_overlay.json . --section all
go run --overlay ../overlays/gpma_overlay.json . --section algorithms
go run --overlay ../overlays/gpma_overlay.json . --section order
```

### Check modes

```bash
CHECK_MODE=console go run --overlay ../overlays/gpma_overlay.json . --section all
CHECK_MODE=silent go run --overlay ../overlays/gpma_overlay.json . --section all
```

## Full demo flow

```bash
./demo/scripts/demo-all.sh --include-failure --run
```

## Debug mode

```bash
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra --plugin-config gobra:debug:true \
  /work/demo/algorithms/math/gcd.go
```

Writes `debug.gobra` into the mounted working directory.

## Alternative: GPMA built from source (development)

If you build the framework locally and have Java + Z3 + gobra.jar:

```bash
cd framework && make build
./build/framework file --plugin gobra ../demo/algorithms/math/gcd.go
```

For local development without native Gobra dependencies, use `gobra:backend:docker` with the host binary — but for demos prefer **`gpma:latest`** as shown above.
