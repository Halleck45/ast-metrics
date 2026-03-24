<p align="center" style="text-align:center">
<img alt="AST Metrics" src="https://raw.githubusercontent.com/Halleck45/ast-metrics/main/docs/logo-ast-metrics-condensed.png" height="200px"/>
</p>

<p align="center" style="text-align:center">
<b>No server. No account. One binary.</b>
<br />
AST Metrics analyzes your codebase (complexity, architecture, coupling, bus factor...) and runs anywhere.
<br />
Drop it in any CI. Works offline. Nothing to install, no SaaS, no data leaves your machine.
</p>
<br />

<p align="center" style="text-align:center">
<a href="https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml"><img src="https://github.com/Halleck45/ast-metrics/actions/workflows/test.yml/badge.svg" alt="CI"></a>
<img src="https://img.shields.io/github/v/release/Halleck45/ast-metrics" alt="GitHub Release">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
<a href="https://github.com/sponsors/Halleck45"><img src="https://img.shields.io/static/v1?label=Sponsor&amp;message=%E2%9D%A4&amp;logo=GitHub&amp;color=%23fe8e86" alt=""></a>
<img src="https://img.shields.io/github/downloads/Halleck45/ast-metrics/total" alt="GitHub all releases">
<a href="https://goreportcard.com/report/github.com/Halleck45/ast-metrics"><img src="https://goreportcard.com/badge/github.com/Halleck45/ast-metrics" alt="Go Report Card"></a>
<a href="https://codecov.io/gh/Halleck45/ast-metrics"><img src="https://codecov.io/gh/Halleck45/ast-metrics/branch/main/graph/badge.svg" alt="codecov"></a>
<a href="https://analyze.ast-metrics.dev/halleck45/ast-metrics"><img src="https://img.shields.io/badge/AST--Metrics-report-181717?logo=github" alt="AST-Metrics report"></a>
</p>

<p align="center" style="text-align:center">
<a href="https://ast-metrics.dev/">Documentation</a> | <a href=".github/CONTRIBUTING.md">Contributing</a> | <a href="https://twitter.com/Halleck45">Twitter</a>
</p>

<img width="1280" height="640" alt="banner" src="https://github.com/user-attachments/assets/4a7d518d-82fe-4c18-880f-479fe1738878" />

<br />

<table align="center">
<tr>
<td align="center">
<br />

<a href="https://analyze.ast-metrics.dev"><img src="https://img.shields.io/badge/Analyze_your_project-4f46e5?style=for-the-badge&logo=data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0id2hpdGUiPjxwYXRoIGQ9Ik0xMiAyQzYuNDggMiAyIDYuNDggMiAxMnM0LjQ4IDEwIDEwIDEwIDEwLTQuNDggMTAtMTBTMTcuNTIgMiAxMiAyem0tMiAxNWwtNS01IDEuNDEtMS40MUwxMCAxNC4xN2w3LjU5LTcuNTlMMTkgOGwtOSA5eiIvPjwvc3ZnPg==&logoColor=white" alt="Analyze your project" height="45"></a>

<p>Paste a GitHub URL. Get a full report. No install.</p>

<p><b>Or explore live examples:</b></p>

<p>
<a href="https://analyze.ast-metrics.dev/spf13/cobra"><img src="https://img.shields.io/badge/spf13%2Fcobra-Go-00ADD8?style=flat-square&logo=go&logoColor=white" alt="spf13/cobra"></a>
<a href="https://analyze.ast-metrics.dev/fatih/color"><img src="https://img.shields.io/badge/fatih%2Fcolor-Go-00ADD8?style=flat-square&logo=go&logoColor=white" alt="fatih/color"></a>
<a href="https://analyze.ast-metrics.dev/gorilla/mux"><img src="https://img.shields.io/badge/gorilla%2Fmux-Go-00ADD8?style=flat-square&logo=go&logoColor=white" alt="gorilla/mux"></a>
<a href="https://analyze.ast-metrics.dev/guzzle/psr7"><img src="https://img.shields.io/badge/guzzle%2Fpsr7-PHP-777BB4?style=flat-square&logo=php&logoColor=white" alt="guzzle/psr7"></a>
<a href="https://analyze.ast-metrics.dev/thephpleague/flysystem"><img src="https://img.shields.io/badge/thephpleague%2Fflysystem-PHP-777BB4?style=flat-square&logo=php&logoColor=white" alt="thephpleague/flysystem"></a>
</p>

<br />

</td>
</tr>
</table>

<br />

## Getting Started

Open your terminal and run the following command:

```console
curl -s https://raw.githubusercontent.com/Halleck45/ast-metrics/main/scripts/download.sh|bash
./ast-metrics analyze --report-html=<directory> /path/to/your/code
```

> To install it manually follow the detailed [installation instructions](https://ast-metrics.dev/getting-started/install/).

## What you get

| | |
|---|---|
| **Architectural analysis** | Community detection, coupling, instability — catch design drift early |
| **Code metrics** | Cyclomatic complexity, maintainability index, lines of code |
| **Activity metrics** | Commit history, bus factor — know who owns what |
| **Linter** | Enforce thresholds on coupling, complexity, LOC per method |
| **CI/CD ready** | GitHub Actions, GitLab CI, any pipeline — exits non-zero on violations |
| **Multiple report formats** | HTML dashboard, JSON, Markdown, SARIF, OpenMetrics |
| **MCP server** | Give AI coding agents architectural awareness via Model Context Protocol |

[Read more in the documentation](https://ast-metrics.dev/)


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


## MCP Server for AI agents

AI coding agents (Claude Code, Cursor, Copilot...) read code linearly and lack architectural awareness. AST Metrics can act as an [MCP server](https://modelcontextprotocol.io/) to give them on-demand access to complexity, coupling, dependency graphs, community detection, risk scoring, and test quality — without reading every file.

```bash
ast-metrics mcp .
```

This starts a stdio MCP server exposing 8 tools:

| Tool | Purpose |
|---|---|
| `analyze_project` | High-level overview: languages, complexity, maintainability, top risks |
| `get_file_metrics` | Detailed metrics for a specific file |
| `find_risky_code` | Files/classes with highest risk scores |
| `find_complex_code` | Functions/classes above a complexity threshold |
| `get_dependencies` | Dependency subgraph around a component |
| `get_coupling` | Afferent/efferent coupling for a component |
| `get_communities` | Architectural community detection and metrics |
| `get_test_quality` | Test isolation, traceability, god tests, orphan classes |

To use it with Claude Code or any MCP-compatible agent, add a `.mcp.json` at your project root:

```json
{
  "mcpServers": {
    "ast-metrics": {
      "command": "ast-metrics",
      "args": ["mcp", "."]
    }
  }
}
```

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

## Support the project

If AST Metrics saved you time, a star goes a long way — it helps other developers discover the tool.

[![Star History Chart](https://api.star-history.com/svg?repos=Halleck45/ast-metrics&type=Date)](https://star-history.com/#Halleck45/ast-metrics&Date)
