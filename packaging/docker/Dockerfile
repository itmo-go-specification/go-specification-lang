# Multi-stage: build .deb, then install into runtime image.
#
# Build:
#   docker build -f packaging/docker/Dockerfile -t gpma:latest .
#
# Run:
#   docker run --rm gpma:latest plugins

# ---- build deb -----------------------------------------------------------
FROM golang:1.25-bookworm AS deb-builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    unzip \
    openjdk-21-jre-headless \
    ca-certificates \
    dpkg \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/google/go-containerregistry/cmd/crane@v0.20.3 \
    && go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.41.2

WORKDIR /src
COPY . .

ARG GPMA_VERSION=0.1.0
ENV GPMA_VERSION=${GPMA_VERSION}

RUN chmod +x packaging/scripts/*.sh packaging/debian/* packaging/files/gobra-wrapper.sh \
    && export PATH="${PATH}:$(go env GOPATH)/bin" \
    && packaging/scripts/build-deb.sh

# ---- runtime -------------------------------------------------------------
FROM ubuntu:22.04 AS runtime

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    openjdk-21-jre-headless \
    && rm -rf /var/lib/apt/lists/*

COPY --from=deb-builder /src/packaging/dist/gpma_*.deb /tmp/gpma.deb
RUN apt-get update \
    && apt-get install -y --no-install-recommends /tmp/gpma.deb \
    && rm /tmp/gpma.deb \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/bin:${PATH}"
ENV LD_LIBRARY_PATH=/usr/lib/gpma/lib

ENTRYPOINT ["gpma"]
CMD ["--help"]
