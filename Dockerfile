#
# docker build . -t Halleck45/ast-metrics:latest
# docker run -it --rm Halleck45/ast-metrics:latest sh
# docker run -it -v .:/src  --rm Halleck45/ast-metrics:latest ast-metrics analyze --report-html=/src/repport /src
#
FROM golang:1.22-alpine AS builder

# Install packages
RUN apk --update --no-cache add \
        build-base \
        make \
        curl \
        git \
# Remove alpine cache
    && rm -rf /var/cache/apk/*

WORKDIR /usr/app

COPY . /usr/app

RUN make build

FROM alpine:latest

LABEL maintainer="Halleck45"
LABEL org.opencontainers.image.source=https://github.com/Halleck45/ast-metrics
LABEL org.opencontainers.image.path="Dockerfile"
LABEL org.opencontainers.image.title="ast-metrics"
LABEL org.opencontainers.image.description="AST Metrics is a blazing-fast static code analyzer. It provides metrics about your code, and helps you to identify potential problems early on."
LABEL org.opencontainers.image.authors="Halleck45"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.documentation="https://github.com/Halleck45/ast-metrics/README.md"

WORKDIR /

COPY --from=builder /usr/app/dist/ast-metrics_linux_amd64_v1 /usr/app/dist/ast-metrics_linux_amd64_v1
COPY --from=builder /usr/app/dist/ast-metrics_linux_arm64 /usr/app/dist/ast-metrics_linux_arm64

RUN arch=$(uname -m) \
  && echo "ARCH=$arch" \
  && case "$arch" in \
    x86_64) \
      cp /usr/app/dist/ast-metrics_linux_amd64_v1/ast-metrics /usr/local/bin/ast-metrics ;; \
    aarch64) \
      cp /usr/app/dist/ast-metrics_linux_arm64/ast-metrics /usr/local/bin/ast-metrics ;; \
    *) \
      echo "Architecture inconnue: $arch" && exit 1 ;; \
  esac

RUN chmod +x /usr/local/bin/ast-metrics
RUN rm -rf /usr/app
