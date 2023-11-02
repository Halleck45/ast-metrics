# Contributing

## Requirements

+ Go 1.21+
+ Make

## Setup

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

```bash
make build
```