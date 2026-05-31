#!/bin/sh
set -a
if [ -f /etc/gpma/gobra.env ]; then
  . /etc/gpma/gobra.env
fi
set +a

export LD_LIBRARY_PATH="${LD_LIBRARY_PATH:-/usr/lib/gpma/lib}"

exec "${GOBRA_JAVA:-java}" ${GOBRA_JAVA_OPTS:--Xss128m} -jar "${GOBRA_JAR:-/usr/lib/gpma/gobra/gobra.jar}" "$@"
