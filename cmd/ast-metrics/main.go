package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/internal/cli"
	"github.com/halleck45/ast-metrics/internal/command"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/halleck45/ast-metrics/internal/engine/golang"
	"github.com/halleck45/ast-metrics/internal/engine/php"
	"github.com/halleck45/ast-metrics/internal/engine/python"
	"github.com/halleck45/ast-metrics/internal/engine/rust"
	"github.com/halleck45/ast-metrics/internal/watcher"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
	cliV2 "github.com/urfave/cli/v2"
)

var (
	// Current version. Managed by goreleaser during build
	// @see https://goreleaser.com/cookbooks/using-main.version/
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {

	logrus.SetLevel(logrus.TraceLevel)

	// Create a temporary directory
	build, err := os.MkdirTemp("", "ast-metrics")
	if err != nil {
		logrus.Error(err)
	}
	defer os.RemoveAll(build)

	// Prepare accepted languages
	runnerPhp := php.PhpRunner{}
	runnerGolang := golang.GolangRunner{}
	runnerPython := python.PythonRunner{}
	runnerRust := rust.RustRunner{}
	runners := []engine.Engine{&runnerPhp, &runnerGolang, &runnerPython, &runnerRust}

	app := &cliV2.App{
		Name:  "ast-metrics",
		Usage: "Static code analysis tool",
		Commands: []*cliV2.Command{
			{
				Name:    "analyze",
				Aliases: []string{"a"},
				Usage:   "Start analyzing the project",
				Flags: []cliV2.Flag{
					&cliV2.BoolFlag{
						Name:     "verbose",
						Aliases:  []string{"v"},
						Usage:    "Enable verbose mode",
						Category: "Global options",
					},
					&cliV2.StringSliceFlag{
						Name:     "exclude",
						Usage:    "Regular expression to exclude files from analysis",
						Category: "File selection",
					},
					&cliV2.BoolFlag{
						Name:     "non-interactive",
						Aliases:  []string{"i"},
						Usage:    "Disable interactive mode",
						Category: "Global options",
					},
					// HTML report
					&cliV2.StringFlag{
						Name:     "report-html",
						Usage:    "Generate an HTML report",
						Category: "Report",
					},
					&cliV2.BoolFlag{
						Name:     "open-html",
						Usage:    "Automatically open HTML report in browser",
						Category: "Report",
					},
					// Markdown report
					&cliV2.StringFlag{
						Name:     "report-markdown",
						Usage:    "Generate an Markdown report file",
						Category: "Report",
					},
					// JSON report
					&cliV2.StringFlag{
						Name:     "report-json",
						Usage:    "Generate a report in JSON format",
						Category: "Report",
					},
					// OpenMetrics report
					// https://github.com/prometheus/OpenMetrics/blob/main/specification/OpenMetrics.md
					&cliV2.StringFlag{
						Name:     "report-openmetrics",
						Usage:    "Generate a report in OpenMetrics format",
						Category: "Report",
					},
					// SARIF report
					&cliV2.StringFlag{
						Name:     "report-sarif",
						Usage:    "Generate a report in SARIF format (2.1.0)",
						Category: "Report",
					},
					// Watch mode
					&cliV2.BoolFlag{
						Name:     "watch",
						Usage:    "Re-run the analysis when files change",
						Category: "Global options",
					},
					// CI mode (alias of --non-interactive, --report-html and --report-markdown)
					&cliV2.BoolFlag{
						Name:     "ci",
						Usage:    "Enable CI mode",
						Category: "Global options",
					},
					// Configuration
					&cliV2.StringFlag{
						Name:     "config",
						Usage:    "Load configuration from file",
						Category: "Configuration",
					},
					// Diff mode (comparaison between current branch and another one or commit)
					&cliV2.StringFlag{
						Name:     "compare-with",
						Usage:    "Compare with another Git branch or commit",
						Category: "Global options",
					},
					// Profiling (with pprof)
					&cliV2.BoolFlag{
						Name:     "profile",
						Usage:    "Generate a profiling reports into files ast-metrics.cpu and ast-metrics.mem",
						Category: "Global options",
					},
				},
				Action: func(cCtx *cliV2.Context) error {

					// get option --verbose
					if cCtx.Bool("verbose") {
						logrus.SetLevel(logrus.DebugLevel)
					}

					// get option --profile
					profile := cCtx.Bool("profile")
					if profile {
						cpufile := "ast-metrics.cpu"
						memfile := "ast-metrics.mem"
						f, err := os.Create(cpufile)
						if err != nil {
							logrus.Fatal("could not create CPU profile: ", err)
						}
						defer f.Close() // error handling omitted for example
						if err := pprof.StartCPUProfile(f); err != nil {
							logrus.Fatal("could not start CPU profile: ", err)
						}
						defer pprof.StopCPUProfile()

						f, err = os.Create(memfile)
						if err != nil {
							logrus.Fatal("could not create memory profile: ", err)
						}
						defer f.Close() // error handling omitted for example
						runtime.GC()    // get up-to-date statistics
						if err := pprof.WriteHeapProfile(f); err != nil {
							logrus.Fatal("could not write memory profile: ", err)
						}
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
					config := configuration.NewConfiguration()

					// Load configuration file
					loader := configuration.NewConfigurationLoader()
					if cCtx.String("config") != "" {
						loader.FilenameToChecks = []string{cCtx.String("config")}
					}

					config, err = loader.Loads(config)
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
						if config.SourcesToAnalyzePath == nil || len(config.SourcesToAnalyzePath) == 0 {
							if isInteractive {
								// we try to ask the user to select a file
								pathsSlice = cli.AskUserToSelectFile()
							}
						}
					} else {
						if len(pathsSlice) == 0 && (config.SourcesToAnalyzePath == nil || len(config.SourcesToAnalyzePath) == 0) {
							pterm.Error.Println("Please provide a path to analyze")
							return nil
						}
						err := config.SetSourcesToAnalyzePath(pathsSlice)
						if err != nil {
							pterm.Error.Println(err.Error())
							return err
						}
					}

					// Exclude patterns
					if config.ExcludePatterns == nil {
						excludePatterns := cCtx.StringSlice("exclude")
						if excludePatterns != nil && len(excludePatterns) > 0 {
							config.SetExcludePatterns(excludePatterns)
						}
					}

					// Reports
					if cCtx.String("report-html") != "" {
						config.Reports.Html = cCtx.String("report-html")
					}
					if cCtx.String("report-markdown") != "" {
						config.Reports.Markdown = cCtx.String("report-markdown")
					}
					if cCtx.String("report-json") != "" {
						config.Reports.Json = cCtx.String("report-json")
					}
					if cCtx.String("report-openmetrics") != "" {
						config.Reports.OpenMetrics = cCtx.String("report-openmetrics")
					}
					if cCtx.String("report-sarif") != "" {
						config.Reports.Sarif = cCtx.String("report-sarif")
					}
					if cCtx.Bool("open-html") {
						config.Reports.OpenHtml = true
					}

					// CI mode
					if cCtx.Bool("ci") {
						pterm.Warning.Println("[DEPRECATION] L'option --ci pour 'analyze' est dépréciée. Utilisez plutôt la commande: ast-metrics ci")
						if config.Reports.Html == "" {
							config.Reports.Html = "ast-metrics-html-report"
						}
						if config.Reports.Markdown == "" {
							config.Reports.Markdown = "ast-metrics-markdown-report.md"
						}
						if config.Reports.Json == "" {
							config.Reports.Json = "ast-metrics-report.json"
						}
						if config.Reports.OpenMetrics == "" {
							// we don't prefix the file with ast-metrics- because "metrics.txt" is a common filename for CI
							// @see https://docs.gitlab.com/ee/ci/testing/metrics_reports.html
							config.Reports.OpenMetrics = "metrics.txt"
						}
						if config.Reports.Sarif == "" {
							config.Reports.Sarif = "ast-metrics-report.sarif"
						}
						isInteractive = false
						pterm.DisableColor()
					}

					// Compare with
					if cCtx.String("compare-with") != "" {
						config.CompareWith = cCtx.String("compare-with")
					}

					// Run command
					command := command.NewAnalyzeCommand(config, outWriter, runners, isInteractive)

					// Watch mode
					config.Watching = cCtx.Bool("watch")
					err = watcher.NewCommandWatcher(config).Start(command)
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
				Action: func(cCtx *cliV2.Context) error {
					// Run command
					config := configuration.NewConfiguration()
					command := command.NewCleanCommand(config.Storage)
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
				Action: func(cCtx *cliV2.Context) error {
					// Run command
					command := command.NewSelfUpdateCommand(version)
					err := command.Execute()
					if err != nil {
						pterm.Error.Println(err.Error())
						return err
					}
					return nil
				},
			},
			{
				Name:  "ruleset",
				Usage: "Manage requirement rulesets",
				Subcommands: []*cliV2.Command{
					{
						Name:  "list",
						Usage: "List available rulesets",
						Action: func(cCtx *cliV2.Context) error {
							command := command.NewRulesetListCommand()
							return command.Execute()
						},
					},
					{
						Name:  "show",
						Usage: "Show rules inside a ruleset",
						Action: func(cCtx *cliV2.Context) error {
							if cCtx.Args().Len() == 0 {
								return fmt.Errorf("usage: ast-metrics ruleset show <name>")
							}
							name := cCtx.Args().First()
							command := command.NewRulesetShowCommand(name)
							return command.Execute()
						},
					},
					{
						Name:  "add",
						Usage: "Add all rules from a ruleset to the configuration file",
						Action: func(cCtx *cliV2.Context) error {
							if cCtx.Args().Len() == 0 {
								return fmt.Errorf("usage: ast-metrics ruleset add <name>")
							}
							name := cCtx.Args().First()
							command := command.NewRulesetAddCommand(name)
							return command.Execute()
						},
					},
				},
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print version information",
				Action: func(cCtx *cliV2.Context) error {
					// Run command
					command := command.NewVersionCommand(version)
					err := command.Execute()
					if err != nil {
						pterm.Error.Println(err.Error())
						return err
					}
					return nil
				},
			},
			{
				Name:    "lint",
				Aliases: []string{"l"},
				Usage:   "Run analysis and print lint (requirements) only",
				Flags: []cliV2.Flag{
					&cliV2.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose mode", Category: "Global options"},
					&cliV2.StringSliceFlag{Name: "exclude", Usage: "Regular expression to exclude files from analysis", Category: "File selection"},
					&cliV2.StringFlag{Name: "config", Usage: "Load configuration from file", Category: "Configuration"},
					&cliV2.StringFlag{Name: "report-sarif", Usage: "Write lint violations as SARIF 2.1.0 to the given file", Category: "Report"},
				},
				Action: func(cCtx *cliV2.Context) error {
					if cCtx.Bool("verbose") {
						logrus.SetLevel(logrus.DebugLevel)
					}
					outWriter := bufio.NewWriter(os.Stdout)
					config := configuration.NewConfiguration()
					loader := configuration.NewConfigurationLoader()
					if cCtx.String("config") != "" {
						loader.FilenameToChecks = []string{cCtx.String("config")}
					}
					cfg, err := loader.Loads(config)
					if err != nil {
						pterm.Error.Println("Cannot load configuration file: " + err.Error())
					}
					// report sarif flag
					if cCtx.String("report-sarif") != "" {
						cfg.Reports.Sarif = cCtx.String("report-sarif")
					}

					// paths from args
					paths := cCtx.Args()
					if paths.Len() > 0 {
						pathsSlice := make([]string, paths.Len())
						for i := 0; i < paths.Len(); i++ {
							pathsSlice[i] = paths.Get(i)
						}
						if err := cfg.SetSourcesToAnalyzePath(pathsSlice); err != nil {
							pterm.Error.Println(err.Error())
							return err
						}
					}
					// exclude
					if cfg.ExcludePatterns == nil {
						ex := cCtx.StringSlice("exclude")
						if len(ex) > 0 {
							cfg.SetExcludePatterns(ex)
						}
					}
					// No report generation here; just lint
					cmd := command.NewLintCommand(cfg, outWriter, runners)
					// pass verbose to command
					cmd.SetVerbose(cCtx.Bool("verbose"))
					command := cmd
					if err := command.Execute(); err != nil {
						return err
					}

					pterm.Success.Println("No lint violations found.")

					return nil
				},
			},
			{
				Name:  "ci",
				Usage: "Run lint then full analysis with reports (CI mode)",
				Flags: []cliV2.Flag{
					&cliV2.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose mode", Category: "Global options"},
					&cliV2.StringSliceFlag{Name: "exclude", Usage: "Regular expression to exclude files from analysis", Category: "File selection"},
					&cliV2.StringFlag{Name: "report-html", Usage: "Generate an HTML report", Category: "Report"},
					&cliV2.BoolFlag{Name: "open-html", Usage: "Automatically open HTML report in browser", Category: "Report"},
					&cliV2.StringFlag{Name: "report-markdown", Usage: "Generate a Markdown report file", Category: "Report"},
					&cliV2.StringFlag{Name: "report-json", Usage: "Generate a report in JSON format", Category: "Report"},
					&cliV2.StringFlag{Name: "report-openmetrics", Usage: "Generate a report in OpenMetrics format", Category: "Report"},
					&cliV2.StringFlag{Name: "report-sarif", Usage: "Generate a report in SARIF format (2.1.0)", Category: "Report"},
					&cliV2.StringFlag{Name: "config", Usage: "Load configuration from file", Category: "Configuration"},
					&cliV2.StringFlag{Name: "compare-with", Usage: "Compare with another Git branch or commit", Category: "Global options"},
				},
				Action: func(cCtx *cliV2.Context) error {
					if cCtx.Bool("verbose") {
						logrus.SetLevel(logrus.DebugLevel)
					}
					// Stdout
					outWriter := bufio.NewWriter(os.Stdout)
					// Prepare configuration object
					config := configuration.NewConfiguration()
					// Load configuration file
					loader := configuration.NewConfigurationLoader()
					if cCtx.String("config") != "" {
						loader.FilenameToChecks = []string{cCtx.String("config")}
					}
					cfg, err := loader.Loads(config)
					if err != nil {
						pterm.Error.Println("Cannot load configuration file: " + err.Error())
					}
					// Paths from args
					paths := cCtx.Args()
					if paths.Len() > 0 {
						pathsSlice := make([]string, paths.Len())
						for i := 0; i < paths.Len(); i++ {
							pathsSlice[i] = paths.Get(i)
						}
						if err := cfg.SetSourcesToAnalyzePath(pathsSlice); err != nil {
							pterm.Error.Println(err.Error())
							return err
						}
					}
					// Exclude patterns
					if cfg.ExcludePatterns == nil {
						excludePatterns := cCtx.StringSlice("exclude")
						if len(excludePatterns) > 0 {
							cfg.SetExcludePatterns(excludePatterns)
						}
					}
					// Reports from flags
					if cCtx.String("report-html") != "" {
						cfg.Reports.Html = cCtx.String("report-html")
					}
					if cCtx.String("report-markdown") != "" {
						cfg.Reports.Markdown = cCtx.String("report-markdown")
					}
					if cCtx.String("report-json") != "" {
						cfg.Reports.Json = cCtx.String("report-json")
					}
					if cCtx.String("report-openmetrics") != "" {
						cfg.Reports.OpenMetrics = cCtx.String("report-openmetrics")
					}
					if cCtx.String("report-sarif") != "" {
						cfg.Reports.Sarif = cCtx.String("report-sarif")
					}
					if cCtx.Bool("open-html") {
						cfg.Reports.OpenHtml = true
					}
					// CI defaults for reports if not set
					if cfg.Reports.Html == "" {
						cfg.Reports.Html = "ast-metrics-html-report"
					}
					if cfg.Reports.Markdown == "" {
						cfg.Reports.Markdown = "ast-metrics-markdown-report.md"
					}
					if cfg.Reports.Json == "" {
						cfg.Reports.Json = "ast-metrics-report.json"
					}
					if cfg.Reports.OpenMetrics == "" {
						cfg.Reports.OpenMetrics = "metrics.txt"
					}
					if cfg.Reports.Sarif == "" {
						cfg.Reports.Sarif = "ast-metrics-report.sarif"
					}
					// Compare with
					if cCtx.String("compare-with") != "" {
						cfg.CompareWith = cCtx.String("compare-with")
					}
					// Run CI command
					cmd := command.NewCICommand(cfg, outWriter, runners)
					if err := cmd.Execute(); err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:  "deploy:github",
				Usage: "Deploy AST-Metrics workflow to all repositories in a GitHub organization. It open a PR for each repository.",
				Flags: []cliV2.Flag{
					&cliV2.StringFlag{
						Name:     "token",
						Usage:    "GitHub personal access token (required)",
						Required: true,
					},
					&cliV2.StringFlag{
						Name:  "branch",
						Usage: "Branch name to create (default: chore/ast-metrics-setup)",
						Value: "chore/ast-metrics-setup",
					},
					&cliV2.StringFlag{
						Name:  "workflow-path",
						Usage: "Path to the workflow file (default: .github/workflows/ast-metrics.yml)",
						Value: ".github/workflows/ast-metrics.yml",
					},
					&cliV2.BoolFlag{
						Name:  "include-forks",
						Usage: "Include forked repositories",
						Value: false,
					},
				},
				Action: func(cCtx *cliV2.Context) error {
					if cCtx.Args().Len() == 0 {
						return fmt.Errorf("usage: ast-metrics deploy:github --token <token> <org>")
					}
					org := cCtx.Args().First()
					token := cCtx.String("token")
					branch := cCtx.String("branch")
					workflowPath := cCtx.String("workflow-path")
					includeForks := cCtx.Bool("include-forks")

					cmd := command.NewDeployGithubOrganizationCommand(org, token, branch, workflowPath, includeForks)
					err := cmd.Execute()
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
				Action: func(cCtx *cliV2.Context) error {
					// Run command
					command := command.NewInitConfigurationCommand()
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
		logrus.Error(err)
	}
}
