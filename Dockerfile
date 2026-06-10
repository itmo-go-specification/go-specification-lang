FROM golang:1.25-trixie

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    git \
    make \
    python3 \
    python3-pip \
    fish \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

    # openjdk-21-jre-headless for framework/gosl/grammar/generate.sh 
    # dpkg, unzip for building deb package inside of this container
RUN apt-get update && apt-get install -y --no-install-recommends \
    openjdk-21-jre-headless \ 
    dpkg \
    unzip \
    && rm -rf /var/lib/apt/lists/*


RUN curl -fsSL https://just.systems/install.sh | bash -s -- --to /usr/local/bin

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /work
CMD ["fish"]

