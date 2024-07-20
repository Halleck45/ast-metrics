# Generate reports

## HTML reports

AST Metrics can generate HTML reports. The reports provide an overview of the codebase, including:

- The number of files and directories
- The number of lines of code
- Maintainability, complexity, and risk scores

To generate a report, run the following command in your terminal:

```bash
ast-metrics --report-html=<report-directory> /path/to/your/project
```

Where `<report-directory>` is the directory where the report will be saved.

## Markdown reports

AST Metrics can also generate Markdown reports. The reports provide an overview of the codebase, in markdown format.

To generate a report, run the following command in your terminal:

```bash
ast-metrics --report-markdown=<report-file.md> /path/to/your/project
```

Where `<report-file.md>` is the file where the report will be saved.
