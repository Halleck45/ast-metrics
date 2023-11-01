# GhosTea Metrics

GhosTea Metrics is a language-agnostic static code analyzer.

![logo](docs/ghostea-small.png)

## Usage

```bash
ghostea analyze <path>
```

## Requirements

Requirements depend on your stack.

**If you have docker installed**, GhosTea downloads automatically 
all required dependencies.

**If you don't have docker installed**, you need to have language you want to analyze installed on your machine. For example, `php` is required if you want to analyze a php project.

In this case, please use the `--driver=native` option.

## Supported languages

For the moment, only Php is supported. But we are working on adding more languages.

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

See [LICENSE](LICENSE).
