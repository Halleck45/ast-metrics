# Contributing

Thank you for your interest, it's amazing 🥰. Here is some information to help you contribute to this project.

I hope these informations will help you:

> - [How to contribute?](#-how-to-contribute)
> - [I have never programmed in Go, but I would love to learn](#-i-have-never-programmed-in-go-but-i-would-love-to-learn)
> - [How to run automated tests?](#-how-to-run-automated-tests)
> - [How is the source code organized?](#-how-is-the-source-code-organized)
> - [I want to add or modify a report](#-i-want-to-add-or-modify-a-report)
> - [My contribution is about supporting a new programming language](#-my-contribution-is-about-supporting-a-new-programming-language)
> - [My contribution involves updating the data structure (protobuf)](#-my-contribution-involves-updating-the-data-structure-protobuf)
> - [How to release new version?](#-how-to-release-new-version)
> - [How to improve the website?](#-how-to-improve-the-website)


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

The `NodeType` structure is the most important part of the project. It's the data structure that will be used to store the AST of the source code, and analysis results. This structure is automatically generated from the `proto/NodeType.proto` file.

> [!INFO]
> 
> An **AST (Abstract Syntax Tree)** is a tree representation of the source code. It's used to represent the structure of the source code in a way that is easy to analyze.
>
> A **Statement** is a node in the AST. It represents a part of the source code, like a function, a class, an if statement, etc.

Each statement in the AST can have another statements as children, recursively. For example, the `StmtClass` statement can have `StmtFunction` as children, which can have `StmtDecisionIf` as children, etc. 

In all case, each statement has at least the following properties:

+ `Name`: the structured name of the statement
+ `Stmts`: a list of children statements
+ `Analyze`: a list of analysis results (Volume, complexity, etc.)
+ (optional) `StmtLocationInFile`: tracks the location of the statement in the source code

The role of each `Engine` consists in parsing the source code, and creating the corresponding `Stmt*` statement(s).

Then, each `Analyzer` will then traverse the AST, and compute analysis results (like cyclomatic complexity, volume, etc.) filling the `Analyze` property of each statement.

Do not hesitate to look at the existing analyzers to understand how to use the `Analyze` property. An analyzer is a structure that implements the `Visitor` interface, and will be used to traverse the AST.

This [visitor pattern](https://en.wikipedia.org/wiki/Visitor_pattern) is used to traverse the AST, and compute analysis results.

```go
type Visitor interface {
	Visit(stmts *pb.Stmts, parents *pb.Stmts)
	LeaveNode(stmts *pb.Stmts)
}
```

You'll find an example of an analyzer in the `src/Analyzer/Complexity/CyclomaticVisitor.go` file.

> [!INFO]
> 
> If you want to discover protobuf, you can read the [official documentation](https://protobuf.dev/).

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