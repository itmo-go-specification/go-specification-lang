container := "gpma-builder"

@build-image:
    docker build -t {{ container }} .

build module dir="build":
    mkdir -p {{ dir }}
    go build -o {{ dir }} ./{{ module }}

container:
    docker run --rm -it -v $(pwd):/work {{ container }}
