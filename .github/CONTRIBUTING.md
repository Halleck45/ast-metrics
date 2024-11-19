# Contributing

Thank you for your interest; it's greatly appreciated 🥰. Here is some information to help you contribute to this project.

I hope this information will help you:

> - [I have never programmed in Go, but I would love to learn](#-i-have-never-programmed-in-go-but-i-would-love-to-learn)
> - [How do I run automated tests?](#-how-to-run-automated-tests)
> - [How is the source code organized?](#-how-is-the-source-code-organized)
> - [Adding or modifying a report](#-i-want-to-add-or-modify-a-report)
> - [Supporting a new programming language](#-my-contribution-is-about-supporting-a-new-programming-language)
> - [Updating the data structure (protobuf)](#-my-contribution-involves-updating-the-data-structure-protobuf)
> - [How to release new version?](#-how-to-release-new-version)
> - [improving the website?](#-how-to-improve-the-website)


## 🤓 I have never programmed in Go, but I would love to learn

No problem! Golang is accessible and easy to learn. Here are some resources to get you started:

+ [A Tour of Go](https://tour.golang.org/welcome/1): an interactive tour that will help you get started with Go.
+ [Go by Example](https://gobyexample.com/): a hands-on introduction to Go using annotated example programs.
+ [Go in 5 minutes](https://www.youtube.com/c/go-in-5-minutes): a series of short videos that will help you get started with Go.

You will need `Go 1.21+` to contribute to the source code. Please follow the [official installation guide](https://go.dev/doc/install) to install Go on your machine.

## 🤖 How to run automated tests?

To run automated tests, use the following command:

```bash
go test ./...
```

## 📂 How is the source code organized?

The main directories of the application are as follows:

+ `src/Analyzer`: contains everything related to AST analysis (complexity, volume, etc.)
+ `src/Configuration`: manages configuration (loading files, validation, etc.)
+ `src/Engine`: contains various engines that convert source code (Python, Golang, PHP...) into an AST
+ `src/Report`: generates reports (HTML, markdown, etc.)

## 📃 I want to add or modify a report

Reports can be generated in formats like HTML, markdown, etc.

To add a new report, you need to create a structure that implements the `Reporter` interface defined in `src/Report/Reporter.go`.

```go
type Reporter interface {
	// generates a report based on the files and the project aggregated data
	Generate(files []*pb.File, projectAggregated Analyzer.ProjectAggregated) ([]GeneratedReport, error)
}
```

Then, register this new `Reporter` in the list of available reporters in the file `src/Report/ReportersFactory.go`.

Finally, add a CLI option (e.g., `--report-myreport=file1.foo`) to activate this report by modifying the `main.go` file.

## 🔥 My contribution is about supporting a new programming language

Language agnosticism in the analysis is achieved by using protobuf files that act as intermediaries between the parsed file (an AST) and the analysis engine.

To add support for a new programming language, you need to declare a new Engine that implements the `Engine` interface defined in `src/Engine/Engine.go`.

```go
type Engine interface {
    // Returns true when analyzed files are concerned by the programming language
	IsRequired() bool

	// Prepare the engine for the analysis. For example, in order to prepare caches
	Ensure() error

	// First step of analysis. Parse all files, and generate protobuff compatible AST files
	DumpAST()

	// Cleanups the engine. For example, to remove caches
	Finish() error

	// Give a UI progress bar to the engine
	SetProgressbar(progressbar *pterm.SpinnerPrinter)

	// Give the configuration to the engine
	SetConfiguration(configuration *Configuration.Configuration)

	// Parse a file and return a protobuff compatible AST object
	Parse(filepath string) (*pb.File, error)
}
```

The protobuf file is defined in the proto directory, and a corresponding Go file is generated with the `make build-protobuff` command. This file is versioned as src/NodeType.NodeType.go.

## 🚩 My contribution involves updating the data structure (protobuf)

The data structure is defined in protobuf files located in the proto directory.

You need to install protobuf by running:

```bash
make install-protobuf
```

Once you've finished editing the protobuf files, you can generate the corresponding Go code with the command:

```bash
make build-protobuff
```

This will also verify that the protobuf files are well-formed.

## 📦 How to release new version?

First ensure tests pass:

```bash
make test
```

Then release new version:

```bash
make build
```

## 🌐 How to improve the website?

The [website](https://halleck45.github.io/ast-metrics/) is hosted on Github, using the `documentation` branch.

When the `documentation` branch is updated, the website is automatically updated. You just need to create a pull request on the `documentation` branch.

You'll need [mkdocs](https://www.mkdocs.org/) to build the website locally. Once installed, you can run the following command to build the website:

```bash
mkdocs serve
```