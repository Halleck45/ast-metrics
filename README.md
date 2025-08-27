# AST Metrics 

<img src="https://raw.githubusercontent.com/Halleck45/ast-metrics/main/docs/logo-ast-metrics-right.jpg" height="200px" alt="PhpMetrics" align="left" style="margin-right:20px"/>

[![CI](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml/badge.svg)](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml)
![GitHub Release](https://img.shields.io/github/v/release/Halleck45/ast-metrics)


AST Metrics is a blazing-fast static code analyzer. It provides metrics about your code, and helps you to identify potential problems early on. 

[Documentation](https://halleck45.github.io/ast-metrics/) | [Contributing](.github/CONTRIBUTING.md) | [Twitter](https://twitter.com/Halleck45)

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

or follow the detailled [installation instructions](https://halleck45.github.io/ast-metrics/getting-started/install/).

> [!IMPORTANT]
> Please always read any script found on the internet before running it, and never use privileged access to run it.

## Why?

Static code analysis is a game-changer for improving code quality. It helps you catch potential issues early, enforce coding standards, and gain deeper insights into your codebase.

With **AST Metrics**, you can:

- **Get a bird's-eye view of your project**: Quickly identify overly complex code or tightly coupled classes that could slow down your development or introduce bugs.  
- **Enforce architectural coherence**: Ensure your code structure aligns with best practices—for instance, making sure your service layer doesn’t depend on your data access layer.

Whether you're maintaining an existing codebase or building a new one, AST Metrics empowers you to write better, more maintainable code.

## Features

+ **Designed for CI/CD**. You can integrate it into your pipeline to check that your code meets your quality standards, You can integrate it into your CI pipelines, whether you're using [GitHub Actions](https://halleck45.github.io/ast-metrics/ci/github-actions/) or [GitLab CI](https://halleck45.github.io/ast-metrics/ci/gitlab-ci/).
+ **Fast and efficient**.
+ Provides simple and detailed reports.
+ **Code analysis**: *cyclomatic complexity, maintainability, size...*
+ **Coupling analysis**: *instability, afferent coupling...*
+ **Activity analysis**: *number of commits, bus factor...*
+ **Easy to install**: *no dependencies, no configuration needed*.

[Read more in the documentation](https://halleck45.github.io/ast-metrics/)

## Contributing

AST Metrics is experimental and actively developed. We welcome contributions.

Feel free to [open a discussion](https://github.com/Halleck45/ast-metrics/discussions). We love suggestions, ideas, bug reports, and other contributions.

If you'd like to contribute to the codebase, **check out the [contributing guide](.github/CONTRIBUTING.md) to get started.**

We are looking for help to support new programming languages, stabilize the tool, and enrich it. Here is the list of supported languages:

+ ✅ **PHP** (PHP 5, PHP 7, =< PHP 8.1)
+ ✅ **Golang**
+ ✅ **Python** (Python 2, Python 3)
+ 🕛 **Dart**
+ 🕛 **Flutter**
+ 🕛 **TypeScript**
+ 🕛 **Java**

## License

AST Metrics is open-source software [licensed under the MIT license](LICENSE)
