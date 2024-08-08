# AST Metrics 

<img src="https://raw.githubusercontent.com/Halleck45/ast-metrics/main/docs/logo-ast-metrics-right.jpg" height="200px" alt="PhpMetrics" align="left" style="margin-right:20px"/>

[![CI](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml/badge.svg)](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml)
![GitHub Release](https://img.shields.io/github/v/release/Halleck45/ast-metrics)


AST Metrics is a blazing-fast static code analyzer. It provides metrics about your code, and helps you to identify potential problems early on. 

[Documentation](https://halleck45.github.io/ast-metrics/) | [Twitter](https://twitter.com/Halleck45)

<br/><br/>
<br/><br/>

## Preview

![HTML report](./docs/preview-html-report.png)

## Quick start

Open your terminal and run the following command:

```console
ast-metrics analyze --report-html=<directory> /path/to/your/code
```

## Installation

AST Metrics is a standalone package. It does not require any other software to be installed.

```console
curl -s https://raw.githubusercontent.com/Halleck45/ast-metrics/main/scripts/download.sh|bash
```

With Go

```console
go install github.com/halleck45/ast-metrics@latest
```

or follow the detailled [installation instructions](https://halleck45.github.io/ast-metrics/getting-started/install/).

> [!IMPORTANT]
> Please always read any script found on the internet before running it, and never use privileged access to run it.

## Features

+ **Designed for CI/CD**. You can integrate it into your pipeline to check that your code meets your quality standards.
+ **Fast and efficient**.
+ Provides simple and detailed reports.
+ **Code analysis**: *cyclomatic complexity, maintainability, size...*
+ **Coupling analysis**: *instability, afferent coupling...*
+ **Activity analysis**: *number of commits, bus factor...*

[Read more in the documentation](https://halleck45.github.io/ast-metrics/)

## Contributing

AST Metrics is experimental and actively developed. We welcome contributions.

**Feel free to [open a discussion](https://github.com/Halleck45/ast-metrics/discussions)**. We love suggestions, ideas, bug reports, and other contributions.

If you want to contribute code, please read the [contributing guidelines](.github/CONTRIBUTING.md) to get started.

We are looking for help to support new programming languages, stabilize the tool, and enrich it. Here is the list of supported languages:

+ ✅ **PHP** (full)
+ 👷 **Python** (partial)
+ 👷 **Golang** (partial)
+ 🕛 **Dart**
+ 🕛 **Flutter**
+ 🕛 **TypeScript**
+ 🕛 **Java**

## License

AST Metrics is open-source software [licensed under the MIT license](LICENSE)
