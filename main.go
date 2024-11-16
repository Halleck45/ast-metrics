package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/src/Cli"
	"github.com/halleck45/ast-metrics/src/Command"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/halleck45/ast-metrics/src/Engine/Golang"
	"github.com/halleck45/ast-metrics/src/Engine/Php"
	"github.com/halleck45/ast-metrics/src/Engine/Python"
	"github.com/halleck45/ast-metrics/src/Watcher"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	// Current version. Managed by goreleaser during build
	// @see https://goreleaser.com/cookbooks/using-main.version/
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	log.SetLevel(log.TraceLevel)

	// Create a temporary directory
	build, err := os.MkdirTemp("", "ast-metrics")
	if err != nil {
		log.Error(err)
	}
	defer os.RemoveAll(build)

	// Prepare accepted languages
	runnerPhp := Php.PhpRunner{}
	runnerGolang := Golang.GolangRunner{}
	runnerPython := Python.PythonRunner{}
	runners := []Engine.Engine{&runnerPhp, &runnerGolang, &runnerPython}

	app := &cli.App{
		Name:  "ast-metrics",
		Usage: "Static code analysis tool",
		Commands: []*cli.Command{
			{
				Name:    "analyze",
				Aliases: []string{"a"},
				Usage:   "Start analyzing the project",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "verbose",
						Aliases:  []string{"v"},
						Usage:    "Enable verbose mode",
						Category: "Global options",
					},
					&cli.StringSliceFlag{
						Name:     "exclude",
						Usage:    "Regular expression to exclude files from analysis",
						Category: "File selection",
					},
					&cli.BoolFlag{
						Name:     "non-interactive",
						Aliases:  []string{"i"},
						Usage:    "Disable interactive mode",
						Category: "Global options",
					},
					// HTML report
					&cli.StringFlag{
						Name:     "report-html",
						Usage:    "Generate an HTML report",
						Category: "Report",
					},
					// Markdown report
					&cli.StringFlag{
						Name:     "report-markdown",
						Usage:    "Generate an Markdown report file",
						Category: "Report",
					},
					// JSON report
					&cli.StringFlag{
						Name:     "report-json",
						Usage:    "Generate a report in JSON format",
						Category: "Report",
					},
					// OpenMetrics report
					// https://github.com/prometheus/OpenMetrics/blob/main/specification/OpenMetrics.md
					&cli.StringFlag{
						Name:     "report-openmetrics",
						Usage:    "Generate a report in OpenMetrics format",
						Category: "Report",
					},
					// Watch mode
					&cli.BoolFlag{
						Name:     "watch",
						Usage:    "Re-run the analysis when files change",
						Category: "Global options",
					},
					// CI mode (alias of --non-interactive, --report-html and --report-markdown)
					&cli.BoolFlag{
						Name:     "ci",
						Usage:    "Enable CI mode",
						Category: "Global options",
					},
					// Configuration
					&cli.StringFlag{
						Name:     "config",
						Usage:    "Load configuration from file",
						Category: "Configuration",
					},
					// Diff mode (comparaison between current branch and another one or commit)
					&cli.StringFlag{
						Name:     "compare-with",
						Usage:    "Compare with another Git branch or commit",
						Category: "Global options",
					},
				},
				Action: func(cCtx *cli.Context) error {

					// get option --verbose
					if cCtx.Bool("verbose") {
						log.SetLevel(log.DebugLevel)
					}

					// get option --non-interactive
					isInteractive := true
					if cCtx.Bool("non-interactive") || cCtx.Bool("ci") {
						pterm.DisableColor()
						isInteractive = false
					}

					// Stdout
					outWriter := bufio.NewWriter(os.Stdout)
					var style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
					fmt.Println(style.Render("\n🦫 AST Metrics is a language-agnostic static code analyzer."))
					fmt.Println("")

					// Prepare configuration object
					configuration := Configuration.NewConfiguration()

					// Load configuration file
					loader := Configuration.NewConfigurationLoader()
					if cCtx.String("config") != "" {
						loader.FilenameToChecks = []string{cCtx.String("config")}
					}

					configuration, err = loader.Loads(configuration)
					if err != nil {
						pterm.Error.Println("Cannot load configuration file: " + err.Error())
					}

					// If no configuration file is found, we ask the user to select a file or take it from arguments

					// If paths are provided in arguments, we use them
					paths := cCtx.Args()
					pathsSlice := make([]string, paths.Len())
					for i := 0; i < paths.Len(); i++ {
						pathsSlice[i] = paths.Get(i)
					}
					if cCtx.Args().Len() == 0 {
						if configuration.SourcesToAnalyzePath == nil || len(configuration.SourcesToAnalyzePath) == 0 {
							if isInteractive {
								// we try to ask the user to select a file
								pathsSlice = Cli.AskUserToSelectFile()
							}
						}
					} else {
						if len(pathsSlice) == 0 && (configuration.SourcesToAnalyzePath == nil || len(configuration.SourcesToAnalyzePath) == 0) {
							pterm.Error.Println("Please provide a path to analyze")
							return nil
						}
						err := configuration.SetSourcesToAnalyzePath(pathsSlice)
						if err != nil {
							pterm.Error.Println(err.Error())
							return err
						}
					}

					// Exclude patterns
					if configuration.ExcludePatterns == nil {
						excludePatterns := cCtx.StringSlice("exclude")
						if excludePatterns != nil && len(excludePatterns) > 0 {
							configuration.SetExcludePatterns(excludePatterns)
						}
					}

					// Reports
					if cCtx.String("report-html") != "" {
						configuration.Reports.Html = cCtx.String("report-html")
					}
					if cCtx.String("report-markdown") != "" {
						configuration.Reports.Markdown = cCtx.String("report-markdown")
					}
					if cCtx.String("report-json") != "" {
						configuration.Reports.Json = cCtx.String("report-json")
					}
					if cCtx.String("report-openmetrics") != "" {
						configuration.Reports.OpenMetrics = cCtx.String("report-openmetrics")
					}

					// CI mode
					if cCtx.Bool("ci") {
						if configuration.Reports.Html == "" {
							configuration.Reports.Html = "ast-metrics-html-report"
						}
						if configuration.Reports.Markdown == "" {
							configuration.Reports.Markdown = "ast-metrics-markdown-report.md"
						}
						if configuration.Reports.Json == "" {
							configuration.Reports.Json = "ast-metrics-report.json"
						}
						if configuration.Reports.OpenMetrics == "" {
							// we don't prefix the file with ast-metrics- because "metrics.txt" is a common filename for CI
							// @see https://docs.gitlab.com/ee/ci/testing/metrics_reports.html
							configuration.Reports.OpenMetrics = "metrics.txt"
						}
					}

					// Compare with
					if cCtx.String("compare-with") != "" {
						configuration.CompareWith = cCtx.String("compare-with")
					}

					// Run command
					command := Command.NewAnalyzeCommand(configuration, outWriter, runners, isInteractive)

					// Watch mode
					configuration.Watching = cCtx.Bool("watch")
					err = Watcher.NewCommandWatcher(configuration).Start(command)
					if err != nil {
						pterm.Error.Println("Cannot watch files: " + err.Error())
					}

					// Execute command
					err = command.Execute()
					if err != nil {
						pterm.Error.Println(err.Error())
						return err
					}

					return nil
				},
			},
			{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   "Clean workdir",
				Action: func(cCtx *cli.Context) error {
					// Run command
					command := Command.NewCleanCommand()
					err := command.Execute()
					if err != nil {
						pterm.Error.Println(err.Error())
						return err
					}
					return nil
				},
			},
			{
				Name:    "self-update",
				Aliases: []string{"u"},
				Usage:   "Update current binary",
				Action: func(cCtx *cli.Context) error {
					// Run command
					command := Command.NewSelfUpdateCommand(version)
					err := command.Execute()
					if err != nil {
						pterm.Error.Println(err.Error())
						return err
					}
					return nil
				},
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print version information",
				Action: func(cCtx *cli.Context) error {
					// Run command
					command := Command.NewVersionCommand(version)
					err := command.Execute()
					if err != nil {
						pterm.Error.Println(err.Error())
						return err
					}
					return nil
				},
			},
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Create a default configuration file",
				Action: func(cCtx *cli.Context) error {
					// Run command
					command := Command.NewInitConfigurationCommand()
					err := command.Execute()
					if err != nil {
						pterm.Error.Println(err.Error())
						return err
					}
					return nil
				},
			},
		},
	}
	app.Suggest = true

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
