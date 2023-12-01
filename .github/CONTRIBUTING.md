# Contributing

## Requirements

+ Go 1.21+
+ Make

## Setup

Install dependencies:

```bash
make install
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