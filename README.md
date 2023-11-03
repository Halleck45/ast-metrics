# AST Metrics

[![Go](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml/badge.svg)](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml)

![](./docs/preview.gif)

<img src="https://raw.githubusercontent.com/Halleck45/ast-metrics/main/docs/ghostea-small.png" height="80px" alt="AST Metrics" align="left" style="margin-right:20px"/>

AST Metrics is a language-agnostic static code analyzer.

[Twitter](https://twitter.com/Halleck45) | [Contributing](.github/CONTRIBUTING.md)

## Usage

```bash
ast-metrics analyze <path>
```

## Requirements

Requirements depend on your stack (use the `--driver` option):

+ `--driver=docker (default)`: **If you have docker installed**, AST-Metrics downloads automatically 
all required dependencies using docker images.

+ `--driver=native`: **If you don't have docker installed**, you need to have language you want to analyze installed on your machine. For example, `php` is required if you want to analyze a php project.

Note that AST Metrics is faster with the `--driver=native` option.

## Installation

Download the latest version of AST Metrics from the [releases page](https://github.com/Halleck45/ast-metrics/releases/tag/v0.0.1-alpha).

For example, on Linux:

```bash
version=0.0.1-alpha # Replace with the latest version
curl -L https://github.com/Halleck45/ast-metrics/releases/download/${version}/ast-metrics_Linux_i386.tar.gz
mv ast-metrics /usr/local/bin/ast-metrics
chmod +x /usr/local/bin/ast-metrics
```

AST Metrics is installable on Linux, macOS and Windows.

## Supported languages

For the moment, only PHP is supported. But we are working on adding more languages.

## Contributing

See [CONTRIBUTING](.github/CONTRIBUTING.md).

## License

See [LICENSE](LICENSE).
