# AST Metrics

AST Metrics is a language-agnostic static code analyzer.

## Usage

```bash
ast-metrics analyze <path>
```

## Requirements

Requirements depend on your stack (or on the `--driver` option):

+ `--driver=docker (default)`: **If you have docker installed**, AST-Metrics downloads automatically 
all required dependencies using docker images.

+ `--driver=native`: **If you don't have docker installed**, you need to have language you want to analyze installed on your machine. For example, `php` is required if you want to analyze a php project.


## Supported languages

For the moment, only PHP is supported. But we are working on adding more languages.

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

See [LICENSE](LICENSE).
