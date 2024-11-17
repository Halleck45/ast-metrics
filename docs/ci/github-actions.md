# Using AST Metrics in Github action

You can easily integrate AST Metrics into your CI/CD pipeline.

a [Github Action](https://github.com/marketplace/actions/ast-metrics-analysis) is available.

Create a `.github/workflows/ast-metrics.yml` file with the following content:

```yaml
name: AST Metrics
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
        - name: AST Metrics
          uses: halleck45/action-ast-metrics@v1.0.2
```

Now every time you push to your repository, AST Metrics will analyze your code.

Reports will be available on the build summary page.

!!! info "Did you know?"

    You can embed directly the AST Metrics report in the web page of your github action, using the `$GITHUB_STEP_SUMMARY` environment variable.

    ```yaml
    name: AST Metrics
    (...)
    steps:
        - name: Adding markdown
          run: cat ast-metrics-report.md >> $GITHUB_STEP_SUMMARY
    ```
