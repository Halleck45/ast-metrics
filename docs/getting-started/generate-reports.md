# Generate reports

## 🌐 HTML report

AST Metrics can generate HTML reports. The reports provide an overview of the codebase, including:

- The number of files and directories
- The number of lines of code
- Maintainability, complexity, and risk scores

To generate a report, run the following command in your terminal:

```bash
ast-metrics --report-html=<report-directory> /path/to/your/project
```

Where `<report-directory>` is the directory where the report will be saved.

## 📄 Markdown report

AST Metrics can also generate Markdown reports. The reports provide an overview of the codebase, in markdown format.

To generate a report, run the following command in your terminal:

```bash
ast-metrics --report-markdown=<report-file.md> /path/to/your/project
```

Where `<report-file.md>` is the file where the report will be saved.

## 📄 JSON report

AST Metrics can also generate JSON reports. The reports provide an overview of the codebase, in JSON format.

To generate a report, run the following command in your terminal:

```bash
ast-metrics --report-json=<report-file.json> /path/to/your/project
```

Where `<report-file.json>` is the file where the report will be saved.

## 📄 SARIF report

AST Metrics can generate [SARIF](https://sarifweb.azurewebsites.net/) (Static Analysis Results Interchange Format) reports. SARIF is a standard format for the output of static analysis tools, widely supported by security and code quality platforms like GitHub Advanced Security, Azure DevOps, and many CI/CD tools.

To generate a SARIF report, run the following command in your terminal:

```bash
ast-metrics --report-sarif=<report-file.sarif> /path/to/your/project
```

Where `<report-file.sarif>` is the file where the report will be saved.

### Use Cases

SARIF reports are particularly useful for:

- **GitHub Code Scanning**: Upload SARIF files to GitHub to display code quality issues directly in pull requests
- **CI/CD Integration**: Many CI/CD platforms support SARIF for automated code quality checks
- **Security Analysis**: SARIF is the standard format for security scanning tools
- **Tool Interoperability**: Share analysis results between different static analysis tools

## 📄 OpenMetrics report (Gitlab CI)

[OpenMetrics](../ci/gitlab-ci.md) is a standard for metrics exposition. AST Metrics can generate OpenMetrics reports, which can be easily integrated into your CI/CD pipeline, like GitLab CI.

To generate an OpenMetrics report, run the following command in your terminal:

```bash
ast-metrics --report-openmetrics=<report-file.openmetrics> /path/to/your/project
```

Where `<report-file.openmetrics>` is the file where the report will be saved.
