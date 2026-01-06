package main

import (
	"fmt"
	"os"

	"github.com/halleck45/ast-metrics/internal/command"
	"github.com/pterm/pterm"
	cliV2 "github.com/urfave/cli/v2"
)

func main() {
	app := &cliV2.App{
		Name:  "ai_dataset",
		Usage: "Generate a CSV dataset for AI training from source code",
		Flags: []cliV2.Flag{
			&cliV2.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    "Output CSV file path (required)",
				Required: true,
			},
			&cliV2.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose output for debugging",
			},
			&cliV2.IntFlag{
				Name:    "max-files",
				Aliases: []string{"m"},
				Usage:   "Maximum number of files to process per language (0 = unlimited, default: 0)",
				Value:   0,
			},
			&cliV2.IntFlag{
				Name:  "concurrency",
				Usage: "Number of concurrent file processors (default: number of CPU cores, use lower value to reduce memory usage)",
				Value: 0,
			},
		},
		ArgsUsage: "<directory>",
		Action: func(cCtx *cliV2.Context) error {
			if cCtx.Args().Len() == 0 {
				return fmt.Errorf("usage: ai_dataset --output <file.csv> <directory>")
			}
			inputPath := cCtx.Args().First()
			outputPath := cCtx.String("output")
			verbose := cCtx.Bool("verbose")
			maxFiles := cCtx.Int("max-files")
			concurrency := cCtx.Int("concurrency")

			cmd := command.NewAIDatasetCommand(inputPath, outputPath, verbose, maxFiles, concurrency)
			err := cmd.Execute()
			if err != nil {
				pterm.Error.Println(err.Error())
				return err
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		pterm.Error.Println(err.Error())
		os.Exit(1)
	}
}
