@build-image:
    docker build -t "gpma-builder" .

build module dir="build":
    mkdir -p {{ dir }}
    go build -o {{ dir }} ./{{ module }}
