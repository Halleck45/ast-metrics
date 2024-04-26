# AST Metrics [![CI](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml/badge.svg)](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml) [![Release](https://github.com/Halleck45/ast-metrics/actions/workflows/release.yml/badge.svg)](https://github.com/Halleck45/ast-metrics/actions/workflows/release.yml) [![CodeQL](https://github.com/Halleck45/ast-metrics/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/Halleck45/ast-metrics/actions/workflows/github-code-scanning/codeql)


| Terminal application | HTML report |
| --- | ---------- |
| ![AST Metrics is a language-agnostic static code analyzer.](./docs/preview.gif) |![HTML report](./docs/preview-html-report.png) |

**AST Metrics is a blazing-fast static code analyzer that works across programming languages..** It empowers you to gain deep insights into your code structure, identify potential problems early on, and improve code quality.  Leveraging the efficiency of Go, AST Metrics delivers exceptional performance for large codebases.

[Twitter](https://twitter.com/Halleck45) | [Contributing](.github/CONTRIBUTING.md) | 
[Getting started](https://halleck45.github.io/ast-metrics/getting-started/)

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
