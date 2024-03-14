package main

import (
	"bufio"
	"os"

	"github.com/halleck45/ast-metrics/src/Cli"
	"github.com/halleck45/ast-metrics/src/Command"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/halleck45/ast-metrics/src/Engine/Golang"
	"github.com/halleck45/ast-metrics/src/Engine/Php"
	"github.com/halleck45/ast-metrics/src/Engine/Python"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
				},
				Action: func(cCtx *cli.Context) error {

					// get option --verbose
					if cCtx.Bool("verbose") {
						log.SetLevel(log.DebugLevel)
					}

					// get option --non-interactive
					isInteractive := true
					if cCtx.Bool("non-interactive") {
						pterm.DisableColor()
						isInteractive = false
					}

					// Stdout
					outWriter := bufio.NewWriter(os.Stdout)
					pterm.DefaultBasicText.Println(pterm.LightMagenta(" AST Metrics ") + "is a language-agnostic static code analyzer.")

					// Prepare configuration object
					configuration := Configuration.NewConfiguration()

					// Validate path selection
					paths := cCtx.Args()
					pathsSlice := make([]string, paths.Len())
					for i := 0; i < paths.Len(); i++ {
						pathsSlice[i] = paths.Get(i)
					}
					if cCtx.Args().Len() == 0 {
						if isInteractive {
							// we try to ask the user to select a file
							pathsSlice = Cli.AskUserToSelectFile()
						}
					}

					if len(pathsSlice) == 0 {
						pterm.Error.Println("Please provide a path to analyze")
						return nil
					}
					err := configuration.SetSourcesToAnalyzePath(pathsSlice)
					if err != nil {
						pterm.Error.Println(err.Error())
						return err
					}

					// Exclude patterns
					excludePatterns := cCtx.StringSlice("exclude")
					if excludePatterns != nil && len(excludePatterns) > 0 {
						configuration.SetExcludePatterns(excludePatterns)
					}

					// Reports
					if cCtx.String("report-html") != "" {
						configuration.HtmlReportPath = cCtx.String("report-html")
					}

					// Run command
					command := Command.NewAnalyzeCommand(configuration, outWriter, runners, isInteractive)
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
		},
	}
	app.Suggest = true

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
