package main

import (
	"bufio"
	"os"

	"github.com/halleck45/ast-metrics/src/Command"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Driver"
	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/halleck45/ast-metrics/src/Engine/Golang"
	"github.com/halleck45/ast-metrics/src/Engine/Php"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {

	log.SetLevel(log.TraceLevel)
	var driverSelected string

	// Prepare accepted languages
	runnerPhp := Php.PhpRunner{}
	runnerGolang := Golang.GolangRunner{}
	runners := []Engine.Engine{&runnerPhp, &runnerGolang}

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
					&cli.StringFlag{
						Name:        "driver",
						Value:       "docker",
						Usage:       "Driver to use (docker or native)",
						Destination: &driverSelected,
						Category:    "Global options",
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

					// Valide args
					if cCtx.Args().Len() == 0 {
						pterm.Error.Println("Please provide a path to analyze")
						return nil
					}

					// Prepare configuration object
					configuration := Configuration.NewConfiguration()

					// Path
					paths := cCtx.Args()
					// make it slice of strings
					pathsSlice := make([]string, paths.Len())
					for i := 0; i < paths.Len(); i++ {
						pathsSlice[i] = paths.Get(i)
					}
					configuration.SetSourcesToAnalyzePath(pathsSlice)

					// Driver
					var driver Driver.Driver
					driver = Driver.Native
					if driverSelected == "docker" {
						driver = Driver.Docker
					}
					configuration.SetDriver(driver)

					// Exclude patterns
					excludePatterns := cCtx.StringSlice("exclude")
					if excludePatterns != nil && len(excludePatterns) > 0 {
						configuration.SetExcludePatterns(excludePatterns)
					}

					// Run command
					command := Command.NewAnalyzeCommand(configuration, outWriter, runners, isInteractive)
					err := command.Execute()
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
		log.Fatal(err)
	}
}
