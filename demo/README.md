# GPMA Demo Project

A curated Go project with classic algorithms and a small order-pricing domain. Each function is annotated with **GPMA** specifications (`/*@ ... */`) for:

- **Gobra** — static deductive verification (`gpma file --plugin gobra`)
- **dynamic_checks** — runtime contract checks via `go run --overlay`

Algorithms are adapted from [TheAlgorithms/Go](https://github.com/TheAlgorithms/Go) (MIT): rewritten to `int` / `[]int` only (no generics) and annotated for GPMA.

## Layout

```
demo/
  algorithms/
    math/       – GCD, Abs, Clamp, Factorial, Fibonacci
    search/     – binary search, linear search
    sort/       – insertion sort, bubble sort
    strings/    – Hamming distance, palindrome check
    dynamic/    – max subarray (Kadane), coin change
    failure/    – intentionally wrong contracts for Gobra demos
  order/
    pricing.go  – unit price, discounts, order totals
    failure/    – wrong discount postcondition
  scripts/      – Docker helpers (verify-gobra.sh, overlay-docker.sh, demo-all.sh)
  main.go       – runnable demo entrypoint
```

## Quick Start

```bash
# Build demo (plain Go, no overlay)
cd demo && go build ./...

# Run demo
go run . --section all
go run . --section algorithms
go run . --section order
```

## Verification with gpma:latest

Build the GPMA Docker image from the repository root:

```bash
just docker-gpma
```

### Gobra (static verification)

```bash
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/math/gcd.go
```

Batch run:

```bash
./demo/scripts/verify-gobra.sh
```

### dynamic_checks (runtime overlay)

```bash
./demo/scripts/overlay-docker.sh
cd demo && go run --overlay ../overlays/gpma_overlay.json . --section all
```

Full command reference: [DEMO_COMMANDS.md](DEMO_COMMANDS.md).

## Correspondence with TheAlgorithms/Go

| demo/ | TheAlgorithms/Go (reference) |
|-------|------------------------------|
| `algorithms/math/gcd.go` | `math/gcd/` |
| `algorithms/math/factorial.go` | `math/factorial/` |
| `algorithms/math/fibonacci.go` | `math/fibonacci/` |
| `algorithms/search/binary.go` | `searches/binarysearch.go` |
| `algorithms/search/linear.go` | `searches/linearsearch.go` |
| `algorithms/sort/insertion.go` | `sort/insertionsort.go` |
| `algorithms/sort/bubble.go` | `sort/bubblesort.go` |
| `algorithms/strings/hamming.go` | `strings/hamming.go` |
| `algorithms/strings/palindrome.go` | `strings/palindrome/` |
| `algorithms/dynamic/maxsubarray.go` | `dynamic/kadane.go` |
| `algorithms/dynamic/coinchange.go` | `dynamic/coinchange.go` |

`order/pricing.go` is original domain logic (not from TheAlgorithms/Go).

## Annotation Conventions

- Named results `(res int)` for Gobra postconditions
- Boolean results as `(ok int)` with `0` / `1`
- `assigns \nothing;` on pure functions; `assigns arr[0 .. len(arr)-1];` when mutating slices
- `int` and `[]int` only — no generics, no goroutines

## Failure Examples

Files under `algorithms/failure/` and `order/failure/` contain deliberately incorrect `ensures` clauses for demonstrating Gobra verification failures:

```bash
docker run --rm -v "$(pwd):/work" -w /work gpma:latest \
  file --plugin gobra /work/demo/algorithms/failure/binary_search_wrong.go
```
