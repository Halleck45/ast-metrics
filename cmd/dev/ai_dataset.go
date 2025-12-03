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
		},
		ArgsUsage: "<directory>",
		Action: func(cCtx *cliV2.Context) error {
			if cCtx.Args().Len() == 0 {
				return fmt.Errorf("usage: ai_dataset --output <file.csv> <directory>")
			}
			inputPath := cCtx.Args().First()
			outputPath := cCtx.String("output")

			cmd := command.NewAIDatasetCommand(inputPath, outputPath)
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
