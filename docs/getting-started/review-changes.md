# Reviewing your changes

Most quality tools greet you with thousands of pre-existing issues. You close the
window and never come back.

`ast-metrics review` takes the opposite approach: it compares your branch with its
base and reports **only what you made worse**. Existing debt is never mentioned. On
a twenty-year-old codebase, the first run is just as actionable as on a new one.

```bash
ast-metrics review
```

No configuration, no baseline to store, no account. AST Metrics checks out the base
version in a temporary worktree, analyzes both, and diffs the results.

## Reading the output

```
AST Metrics: quality gate passed

24 file(s) changed, 0 new critical issue(s), 17 other regression(s), 5 improvement(s)

Regressions:
- [MEDIUM] internal/engine/php/php_halstead_test.go (internal/engine/php/php_halstead_test.go:47)
      LOC too high in method TestPhpAccessorsAreNotPerfectlyMaintainable(): got 33 (max: 30)
- [MEDIUM] internal/engine/php/php_halstead_test.go (internal/engine/php/php_halstead_test.go:16)
      Function/method name 'TestPhpOperatorsOfAPlainReturn()' contains package name 'php'
  ... and 15 more (see JSON or SARIF report)

Improvements:
- internal/engine/csharp/tree_sitter_csharp_adapter.go: Maintainability index: 43 -> 50
- TreeSitterAdapter::ExtractOperatorsOperands (...:398): Cyclomatic complexity: 14 -> 2

Existing debt is not reported. Methodology v1.0
```

Three things are worth noting:

- every finding names a **file and a line**, so it is actionable without opening the
  report;
- **improvements are reported too**. A review that only ever says "you made things
  worse" is a review people learn to ignore;
- by default the command **never fails**. It informs. You decide when to make it
  blocking.

## Choosing the base

Without `--base`, AST Metrics tries `origin/main`, `origin/master`, `main`, then
`master`. Pass it explicitly when your default branch differs, or to review against
any branch, tag or commit:

```bash
ast-metrics review --base=develop
ast-metrics review --base=HEAD~5
```

## Making it blocking

Once your team trusts the signal, turn the review into a gate with `--fail-on`:

```bash
ast-metrics review --fail-on=high
```

| Value | The command fails when... |
|---|---|
| `never` (default) | never. The review only informs. |
| `high` | a high-severity regression appears. |
| `medium` | a medium or high regression appears. |
| `any` | any regression appears, including low ones. |

Start with `never` for a couple of weeks, look at what the tool actually reports on
your real pull requests, then raise the bar. Turning on `any` from day one is the
surest way to have the check disabled a month later.

## Other output formats

```bash
ast-metrics review --format=markdown          # ready to paste in a pull request
ast-metrics review --report-json=review.json  # the full result, nothing truncated
ast-metrics review --report-sarif=review.sarif
```

The text and Markdown outputs show the five most important regressions; raise the
limit with `--max-findings`, or read the JSON report for the complete list.

!!! tip "Architecture rules count too"
    If your project has an [`.ast-metrics.yaml`](../ci/linting-architecture.md) with
    requirements (forbidden dependencies, complexity budgets, and so on), the review
    also reports the **new** violations your branch introduces, and only those.

## In your pipeline

This is exactly what runs on your pull requests when you install the
[GitHub Action](../ci/github-actions.md), which needs a single line of YAML. See also
[GitLab CI](../ci/gitlab-ci.md).
