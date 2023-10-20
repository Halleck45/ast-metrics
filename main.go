package main

import (
    "embed"
    "log"
    "bufio"
    "os"
    "github.com/urfave/cli/v2"
    "ast-metrics/src/hal/ast-metrics/components/Php"
    "github.com/pterm/pterm"
    "ast-metrics/src/hal/ast-metrics/components/Storage"
)

//go:embed runner/php/vendor/*
var enginPhpSources embed.FS

func main() {

    app := &cli.App{
        Commands: []*cli.Command{
            {
                Name:    "analyze",
                Aliases: []string{"a"},
                Usage:   "Start analyzing the project",
                Action: func(cCtx *cli.Context) error {

                    // Cli app
                    outWriter := bufio.NewWriter(os.Stdout)
                    pterm.DefaultBasicText.Println(pterm.LightMagenta(" AST Metrics ") + "is a language-agnostic static code analyzer.")

                    // Valide args
                    if cCtx.Args().Len() == 0 {
                        pterm.Error.Println("Please provide a path to analyze")
                        return nil
                    }

                    // Prepare workdir
                    Storage.Ensure()

                    // Prepare progress bars
                    multi := pterm.DefaultMultiPrinter.WithWriter(outWriter)
                    spinnerLiveText, _ := pterm.DefaultSpinner.Start("Analyzing project...")

                    progressBarGlobal := pterm.DefaultMultiPrinter
                    pb1, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Checking PHP Engine")
                    pb2, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Parsing PHP files")
                    multi.Start()

                    // Ensure engines are installed
                    // @todo: make engine dynamic, and loop through all engines
                    _, err := Php.Ensure(pb1, enginPhpSources)
                    if err != nil {
                        pterm.Error.Println(err.Error())
                        return err
                    }


                    // Dump ASTs (in parallel)
                    done := make(chan struct{})
                        go func() {
                            path := cCtx.Args().First()
                            Php.DumpAST(pb2, path)
                            close(done)
                        }()
                    <-done

                    // Cleaning up
                    // @todo: make engine dynamic, and loop through all engines
                    _, err = Php.Finish(pb1, enginPhpSources)
                    if err != nil {
                        pterm.Error.Println(err.Error())
                        return err
                    }

                    // Wait for all sub-processes to finish
                    outWriter.Flush()
                    progressBarGlobal.Stop()

                    spinnerLiveText.Success("Finished")
                    return nil
                },
            },
            {
                Name:    "clean",
                Aliases: []string{"i"},
                Usage:   "Clean workdir",
                Action: func(cCtx *cli.Context) error {
                    Storage.Purge()
                    pterm.Success.Println("Workdir cleaned")
                    return nil
                },
            },
        },
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}