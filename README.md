<h1><img alt="AST Metrics" src="https://raw.githubusercontent.com/Halleck45/ast-metrics/main/docs/logo-condensed.png" height="180px" align="left" style="margin-right:40px"/></h1>

[![CI](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml/badge.svg)](https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml)
![GitHub Release](https://img.shields.io/github/v/release/Halleck45/ast-metrics)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/Halleck45)


AST Metrics is a **multi-language static code analyzer**.  

It provides **architectural insights**, **complexity metrics**, and **activity analysis**—all in a **fast, standalone binary** ready for CI/CD.

[Documentation](https://halleck45.github.io/ast-metrics/) | [Contributing](.github/CONTRIBUTING.md) | [Twitter](https://twitter.com/Halleck45)

<br/><br/>
<br/><br/>

## Preview

<table>
    <tr>
        <td width="50%" style="text-align:center">
            HTML Report
        </td>
        <td width="50%" style="text-align:center">
            CLI
        </td>
    </tr>
    <tr>
        <td width="50%" style="text-align:center">
            <img src="./docs/preview-ast-metrics.gif" alt="AST Metrics HTML report"/>
        </td>
        <td width="50%" style="text-align:center">
            <img src="./docs/preview.gif" alt="AST Metrics CLI report"/>
        </td>
    </tr>
</table>



## Quick start

Open your terminal and run the following command:

```console
curl -s https://raw.githubusercontent.com/Halleck45/ast-metrics/main/scripts/download.sh|bash

./ast-metrics analyze --report-html=<directory> /path/to/your/code
```

> AST Metrics is a standalone package. It does not require any other software to be installed.
> To install it manually follow the detailled [installation instructions](https://halleck45.github.io/ast-metrics/getting-started/install/).

## Why AST Metrics?

- **Catch issues early**: detect complex or tightly coupled code.
- **Ensure architectural coherence**: validate dependencies and layering.
- **Understand your project at scale**: from cyclomatic complexity to bus factor.

## Features

+ **CI/CD ready**: plug into GitHub Actions, GitLab CI, or any pipeline.
+ **Fast & dependency-free**: single binary, no setup required.
+ **Architectural analysis**: community detection, coupling, instability.
+ **Code metrics**: complexity, maintainability, size.
+ **Activity metrics**: commits, bus factor.
+ **Readable reports**: detailed HTML dashboards.

[Read more in the documentation](https://halleck45.github.io/ast-metrics/)

## Supported languages

+ ✅ **PHP** `<= PHP 8.4`
+ ✅ **Golang** `any version`
+ ✅ **Python** `Python 2, Python 3`
+ ✅ **Rust** `any version`
+ 🕛 **Dart**
+ 🕛 **Flutter**
+ 🕛 **TypeScript**
+ 🕛 **Java**

## Rule sets: validate your architecture automatically

AST Metrics supports **rulesets**  
You can declare thresholds in your YAML config (Lines of code, Logical lines of code, Coupling, Maintainability...) and AST-Metrics will **fail or succeed the build automatically**.

Example:

```yaml
requirements:
  rules:
    volume:
      loc: { max: 500 }
      lloc_by_method: { max: 30 }
    architecture:
      efferent_coupling: { max: 20 }
      maintainability: { min: 60 }
      coupling:
        forbidden:
          - from: Service
            to: Controller
```

This makes it **easy to enforce architecture and quality at scale**.

Run `ast-metrics ruleset list` to see the list of available rulesets. Then `ast-metrics ruleset add <ruleset-name>` to apply a ruleset to your project.

## License

AST Metrics is open-source software [licensed under the MIT license](LICENSE)


## Contributing

AST Metrics is an actively evolving project.

We welcome discussions, bug reports, and pull requests.

➡️ Start [contributing here](.github/CONTRIBUTING.md)
