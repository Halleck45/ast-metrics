# Tips for your CI

## Generate all reports easily

AST Metrics provides a simple way to integrate code quality metrics into your CI/CD pipeline, using the ` --ci` flag. This flag generates all available reports (HTML, JSON, Markdown and OpenMetrics).

```bash
ast-metrics --ci .
```

## Deploy to multiple repositories at once

If you manage multiple repositories in a GitHub organization, you can deploy AST Metrics to all (or some) of them with a single command. See the [Deploy to GitHub Organization](./deploy-github-org.md) guide for details.

```bash
ast-metrics deploy:github --token=<github-token> <organization-name>
```

## Comparing with another branch

You can compare the metrics of the current branch with another branch using the [`--compare-with`](../advanced-usage/compare-versions.md) flag.

```bash
ast-metrics --ci --compare-with=main .
```
