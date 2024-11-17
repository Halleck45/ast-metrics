# Tips for your CI

## Generate all reports easily

AST Metrics provides a simple way to integrate code quality metrics into your CI/CD pipeline, using the ` --ci` flag. This flag generates all available reports (HTML, JSON, Markdown and OpenMetrics).

```bash
ast-metrics --ci .
```

## Comparing with another branch

You can compare the metrics of the current branch with another branch using the [`--compare-with`](../advanced-usage/compare-versions.md) flag.

```bash
ast-metrics --ci --compare-with=main .
```
