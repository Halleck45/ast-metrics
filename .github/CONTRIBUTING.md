# Contributing

## Requirements

+ Go 1.21+
+ Make
+ Cmake

## Setup

Install dependencies:

```bash
make install
```

AST-Metrics uses C librairies in order to compute metrics. Today, libgit2 is used to parse git repositories.

Go libraries are used to interact with libgit2, but require a specific version of libgit2 to be installed on the system.

In order to install it, please run:

```bash
make install-build
```

## Sources 

Statement descriptors are centralized in protobuf files, in the `proto` directory.

When ready to generate the Go code, run:

```bash
make build-protobuff
```

## Releasing

First ensure tests pass:

```bash
make test
```

Then release new version:

```bash
make build
```