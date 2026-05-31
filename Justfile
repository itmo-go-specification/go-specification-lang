container := "gpma-builder"
gpma_version := "0.1.0"
gpma_image := "gpma"

@fmt:
    gofmt -w .

@build-image:
    docker build -t {{ container }} .

# Build .deb locally (requires Go, curl, unzip; Docker or crane for Gobra JAR).
build-deb version=gpma_version:
    GPMA_VERSION={{ version }} packaging/scripts/build-deb.sh

# Docker image with gpma installed. Builds the .deb inside Docker (no local deb required).
docker-gpma version=gpma_version image=gpma_image:
    docker build -f packaging/docker/Dockerfile \
        --build-arg GPMA_VERSION={{ version }} \
        -t {{ image }}:latest \
        -t {{ image }}:{{ version }} \
        .

# Faster rebuild: builds .deb locally, then installs it into Docker.
docker-gpma-runtime version=gpma_version image=gpma_image: build-deb
    docker build -f packaging/docker/Dockerfile.runtime \
        -t {{ image }}:latest \
        .

build module dir="build":
    mkdir -p {{ dir }}
    go build -o {{ dir }} ./{{ module }}

container:
    docker run --rm -it -v $(pwd):/work {{ container }}
