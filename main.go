package main

import (
    "embed"
    "path/filepath"
    "bufio"
    "os"
    "github.com/urfave/cli/v2"
    "github.com/halleck45/ast-metrics/src/Php"
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Storage"
    "github.com/halleck45/ast-metrics/src/Analyzer"
    log "github.com/sirupsen/logrus"
)

//go:embed engine/php/vendor/* engine/php/generated/* engine/php/dump.php
var enginPhpSources embed.FS

func main() {


    //log.SetLevel(log.TraceLevel)
    log.SetLevel(log.TraceLevel)

    app := &cli.App{

        Commands: []*cli.Command{
            {
                Name:    "analyze",
                Aliases: []string{"a"},
                Usage:   "Start analyzing the project",
                Flags: []cli.Flag{
                    &cli.BoolFlag{
                        Name:  "verbose",
                        Aliases:  []string{"v"},
                        Usage: "Enable verbose mode",
                    },
                },
                Action: func(cCtx *cli.Context) error {

                    // get option --verbose
                    if cCtx.Bool("verbose") {
                        log.SetLevel(log.DebugLevel)
                    }

                    // Cli app
                    outWriter := bufio.NewWriter(os.Stdout)
                    pterm.DefaultBasicText.Println(pterm.LightMagenta(" AST Metrics ") + "is a language-agnostic static code analyzer.")

                    // Valide args
                    if cCtx.Args().Len() == 0 {
                        pterm.Error.Println("Please provide a path to analyze")
                        return nil
                    }

                    path := cCtx.Args().First()
                    // make path absolute
                    if !filepath.IsAbs(path) {
                        var err error
                        path, err = filepath.Abs(path)
                        if err != nil {
                            pterm.Error.Println(err.Error())
                        }
                    }

                    // Prepare workdir
                    Storage.Ensure()

                    // Prepare progress bars
                    multi := pterm.DefaultMultiPrinter.WithWriter(outWriter)
                    spinnerAllExecution, _ := pterm.DefaultProgressbar.WithTotal(3).WithWriter(multi.NewWriter()).WithTitle("Analyzing").Start()

                    pb1, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Checking PHP Engine")
                    pb2, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Parsing PHP files")
                    pbAnalaysis1, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Main analysis")
                    multi.Start()

                    // Ensure engines are installed
                    // @todo: make engine dynamic, and loop through all engines
                    spinnerAllExecution.UpdateTitle("Downloading dependencies...")
                    spinnerAllExecution.Increment()
                    _, err := Php.Ensure(pb1, enginPhpSources, path)
                    if err != nil {
                        pterm.Error.Println(err.Error())
                        return err
                    }

                    // Dump ASTs (in parallel)
                    spinnerAllExecution.UpdateTitle("Dumping AST code...")
                    spinnerAllExecution.Increment()
                    done := make(chan struct{})
                        go func() {
                            Php.DumpAST(pb2, path)
                            close(done)
                        }()
                    <-done


                    // Cleaning up
                    // @todo: make engine dynamic, and loop through all engines
                    _, err = Php.Finish(pb2, enginPhpSources)
                    if err != nil {
                        pterm.Error.Println(err.Error())
                        return err
                    }

                    // Wait for all sub-processes to finish
                    outWriter.Flush()

                    // Now we start the analysis of each AST file
                    spinnerAllExecution.UpdateTitle("Analyzing...")
                    spinnerAllExecution.Increment()
                    Analyzer.Start(pbAnalaysis1)
                    // Start aggregating results
                    pbAnalaysis1.Info("Aggregated results")

                    spinnerAllExecution.UpdateTitle("")
                    spinnerAllExecution.Stop()
                    multi.Stop()

                    // Inform user
                    pterm.DefaultBasicText.Println("")
                    pterm.DefaultBasicText.Println("Finished.")
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