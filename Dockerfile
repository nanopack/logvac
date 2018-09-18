FROM golang:1.10

RUN apt-get update && \
    apt-get install -y \
      # build tools, for compiling
      build-essential \
      # install curl to fetch dev things
      curl \
      # we'll need git for fetching golang deps
      git \
      # we need aws-cli to publish
      awscli

# install dep (not using it yet, but probably will switch to it)
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# setup the app dir/working directory
RUN mkdir -p /go/src/github.com/nanopack/logvac
WORKDIR /go/src/github.com/nanopack/logvac
