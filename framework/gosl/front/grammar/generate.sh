#!/bin/bash
set -e

cd "$(dirname "$0")"

ANTLR_JAR="antlr-4.13.1-complete.jar"
ANTLR_ARGS=(-Xmx500M -cp "$ANTLR_JAR" org.antlr.v4.Tool -Dlanguage=Go -visitor -package grammar)

for grammar in ACSL C; do
    java "${ANTLR_ARGS[@]}" "$grammar.g4" -o "../generated/grammar/${grammar,,}"
done
