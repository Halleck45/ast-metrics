<p align="center" style="text-align:center">
<img alt="AST Metrics" src="https://raw.githubusercontent.com/Halleck45/ast-metrics/main/docs/logo-ast-metrics-condensed.png" height="200px"/>
</p>

<p align="center" style="text-align:center">
AST Metrics is a <b>multi-language static code analyzer</b>.  
<br />
It provides <b>architectural insights</b>, <b>complexity metrics</b>, and <b>activity analysis</b> - all in a <b>fast, standalone binary</b> ready for CI/CD.
</p>

<p align="center" style="text-align:center">
<a href="https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml"><img src="https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml/badge.svg" alt="CI"></a>
<img src="https://img.shields.io/github/v/release/Halleck45/ast-metrics" alt="GitHub Release">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
<a href="https://github.com/sponsors/Halleck45"><img src="https://img.shields.io/static/v1?label=Sponsor&amp;message=%E2%9D%A4&amp;logo=GitHub&amp;color=%23fe8e86" alt=""></a>
<img src="https://img.shields.io/github/downloads/Halleck45/ast-metrics/total" alt="GitHub all releases">
<a href="https://goreportcard.com/report/github.com/Halleck45/ast-metrics"><img src="https://goreportcard.com/badge/github.com/Halleck45/ast-metrics" alt="Go Report Card"></a>
<a href="https://codecov.io/gh/Halleck45/ast-metrics"><img src="https://codecov.io/gh/Halleck45/ast-metrics/branch/main/graph/badge.svg" alt="codecov"></a></p>
</p>

<p align="center" style="text-align:center">
<a href="https://halleck45.github.io/ast-metrics/">Documentation</a> | <a href=".github/CONTRIBUTING.md">Contributing</a> | <a href="https://twitter.com/Halleck45">Twitter</a>
</p>

<br /><br/>

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



## Getting Started

Open your terminal and run the following command:

```console
curl -s https://raw.githubusercontent.com/Halleck45/ast-metrics/main/scripts/download.sh|bash
./ast-metrics analyze --report-html=<directory> /path/to/your/code
```

> To install it manually follow the detailed [installation instructions](https://halleck45.github.io/ast-metrics/getting-started/install/).


## Linting your code

Run:

```bash
# create a .ast-metrics.yaml config file
ast-metrics init 

# Add ruleset to your config file
ast-metrics ruleset add architecture
ast-metrics ruleset add volume
ast-metrics ruleset list # see the list of available rulesets

# Run the linter
ast-metrics lint
```

You can declare thresholds in your YAML config (*Lines of code per method, Coupling, Maintainability...*).

Example:

```yaml
requirements:
  rules:
    architecture:
      coupling:
        forbidden:
          - from: Controller
            to: Repository
          - from: Repository
            to: Service
      max_afferent_coupling: 10
      max_efferent_coupling: 10
      min_maintainability: 70
    volume:
      max_loc: 1000
      max_logical_loc: 600
      max_loc_by_method: 30
      max_logical_loc_by_method: 20
    complexity:
      max_cyclomatic: 10
    golang:
      no_package_name_in_method: true
      max_nesting: 4
      max_file_size: 1000
      max_files_per_package: 50
      slice_prealloc: true
      ignored_error: true
      context_missing: true
      context_ignored: true
```

This makes it **easy to enforce architecture and quality at scale**.

Run `ast-metrics ruleset list` to see the list of available rulesets. Then `ast-metrics ruleset add <ruleset-name>` to apply a ruleset to your project.

## CI usage

Use the dedicated CI command to run lint and generate all reports in one go:

```bash
ast-metrics ci [options] /path/to/your/code
```

Notes:
- This command runs the linter first, then generates HTML, Markdown, JSON, OpenMetrics and SARIF reports.
- If any lint violations are found, the command exits with a non-zero status but still produces the reports.
- The previous alias `analyze --ci` is deprecated and will display a warning. Please migrate to `ast-metrics ci`.

## Github Action

Create a `.github/workflows/ast-metrics.yml` file in your project with the following content:

```yaml
name: "AST Metrics"
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
        - uses: halleck45/action-ast-metrics@v1
```

## Why AST Metrics?

- **Catch issues early**: detect complex or tightly coupled code.
- **Ensure architectural coherence**: validate dependencies and layering.
- **Understand your project at scale**: from cyclomatic complexity to bus factor.

## Features

+ **Architectural analysis**: community detection, coupling, instability.
+ **Linter**: enforce coding standards and best practices.
+ **CI/CD ready**: plug into GitHub Actions, GitLab CI, or any pipeline.
+ **Fast & dependency-free**: single binary, no setup required.
+ **Code metrics**: complexity, maintainability, size.
+ **Activity metrics**: commits, bus factor.
+ **Readable reports**: detailed HTML dashboards.

[Read more in the documentation](https://halleck45.github.io/ast-metrics/)

## Supported languages

+ ✅ **Golang** `any version`
+ ✅ **Python** `Python 2, Python 3`
+ ✅ **Rust** `any version`
+ ✅ **PHP** `<= PHP 8.5`
+ 🕛 **TypeScript**
+ 🕛 **Flutter**
+ 🕛 **Java**
+ 🕛 **C++**
+ 🕛 **Ruby**

## License

AST Metrics is open-source software [licensed under the MIT license](LICENSE)


## Contributing

AST Metrics is an actively evolving project.

We welcome discussions, bug reports, and pull requests.

➡️ Start [contributing here](.github/CONTRIBUTING.md)
