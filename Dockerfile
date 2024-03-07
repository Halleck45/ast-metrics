ARG GOLANG_CROSS_VERSION=v1.22.0
FROM ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION}

RUN apt update && apt install -y pkg-config
RUN apt install -y libgit2-dev
RUN apt install -y libgit2-1.1