# Using AST Metrics in GitLab CI

AST Metrics is compatible with the [OpenMetrics](https://github.com/prometheus/OpenMetrics/blob/main/specification/OpenMetrics.md) standard. This means that you can easily integrate AST Metrics into your GitLab CI/CD pipeline.

Create a `.gitlab-ci.yml` file with the following content:

```yaml
stages:
  - test

test:
    stage: test
    image: ubuntu:latest
    script:
        - curl -s https://raw.githubusercontent.com/Halleck45/ast-metrics/main/scripts/download.sh|sh
        - ./ast-metrics -f --report-openmetrics=metrics.txt .
```

This configuration downloads the latest version of AST Metrics and generates an OpenMetrics report for the current directory. This report is saved in the `metrics.txt` file, and will be available as a [metrics report](https://docs.gitlab.com/ee/ci/testing/metrics_reports.html) in GitLab.