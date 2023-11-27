# AST Metrics

[![Go](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml/badge.svg)](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml)

![](./docs/preview.gif)

<img src="https://github.com/Halleck45/ast-metrics/blob/main/docs/logo-small.png?raw=true" height="80px" alt="AST Metrics" align="left" style="margin-right:20px"/>

AST Metrics is a language-agnostic static code analyzer.

[Twitter](https://twitter.com/Halleck45) | [Contributing](.github/CONTRIBUTING.md)

## Usage

```bash
ast-metrics analyze <path>
```

## Requirements

AST Metrics is a standalone package. It does not require any other software to be installed.

## Installation

Download the latest version of AST Metrics from the [releases page](https://github.com/Halleck45/ast-metrics/releases/tag/v0.0.1-alpha).

For example, on Linux:

```bash
export version=v0.0.1-gamma # Replace with the latest version
curl -L https://github.com/Halleck45/ast-metrics/releases/download/${version}/ast-metrics_Linux_i386.tar.gz -o ast-metrics_Linux_i386.tar.gz
tar -xvf  ast-metrics_Linux_i386.tar.gz
mv ast-metrics /usr/local/bin/ast-metrics
chmod +x /usr/local/bin/ast-metrics
```

AST Metrics is installable on Linux, macOS and Windows.

## Supported languages

For the moment PHP, Python Golang are supported. But we are working on adding more languages.

## Contributing

AST Metrics is experimental and actively developed. We welcome contributions.

See [CONTRIBUTING](.github/CONTRIBUTING.md).

## License

See [LICENSE](LICENSE).
