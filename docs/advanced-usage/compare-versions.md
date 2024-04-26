# Comparing two versions of your code

When you are working on a project, you may want to compare two versions of your code to see how the complexity has evolved over time.

With AST Metrics, you can use the `--compare-with` option to compare two versions of your code.

```console
ast-metrics analyze --compare-with=main
```

This command will compare the current branch with the `main` branch. You can replace `main` with any branch, tag, or commit hash.

