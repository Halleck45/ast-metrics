package main

import (
    "path/filepath"
    "bufio"
    "os"
    "strconv"
    "fmt"
    "github.com/urfave/cli/v2"
    "github.com/halleck45/ast-metrics/src/Php"
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Storage"
    "github.com/halleck45/ast-metrics/src/Engine"
    "github.com/halleck45/ast-metrics/src/Analyzer"
    "github.com/halleck45/ast-metrics/src/Driver"
    log "github.com/sirupsen/logrus"
)

func main() {

    log.SetLevel(log.TraceLevel)
    var driverSelected string

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
                    &cli.StringFlag{
                        Name:        "driver",
                        Value:       "docker",
                        Usage:       "Driver to use (docker or native)",
                        Destination: &driverSelected,
                    },
                },
                Action: func(cCtx *cli.Context) error {

                    // get option --verbose
                    if cCtx.Bool("verbose") {
                        log.SetLevel(log.DebugLevel)
                    }

                    // Cli app
                    outWriter := bufio.NewWriter(os.Stdout)
                    pterm.DefaultBasicText.Println(pterm.LightMagenta(" GhosTea Metrics ") + "is a language-agnostic static code analyzer.")

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
                            return err
                        }
                    }

                    // ensure path exists
                    if _, err := os.Stat(path); err != nil {
                       pterm.Error.Println("Path '" + path + "' does not exist or is not readable")
                       return err
                    }

                    // Prepare workdir
                    Storage.Ensure()

                    // Prepare progress bars
                    multi := pterm.DefaultMultiPrinter.WithWriter(outWriter)
                    spinnerAllExecution, _ := pterm.DefaultProgressbar.WithTotal(3).WithWriter(multi.NewWriter()).WithTitle("Analyzing").Start()

                    // Supported engines are here
                    runnerPhp := Php.PhpRunner{}
                    runners := []Engine.Engine{&runnerPhp}

                    // Start progress bars
                    multi.Start()

                    for _, runner := range runners {

                        // Driver
                        var driver Driver.Driver
                        driver = Driver.Native
                        if driverSelected == "docker" {
                            driver = Driver.Docker
                        }
                        runner.SetDriver(driver)
                        progressBarSpecificForEngine, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Checking PHP Engine")
                        runnerPhp.SetProgressbar(progressBarSpecificForEngine)
                        runnerPhp.SetSourcesToAnalyzePath(path)

                        if runner.IsRequired() {
                            spinnerAllExecution.Increment()
                            err := runner.Ensure()
                            if err != nil {
                                pterm.Error.Println(err.Error())
                                return err
                            }

                            // Dump ASTs (in parallel)
                            spinnerAllExecution.UpdateTitle("Dumping AST code...")
                            spinnerAllExecution.Increment()
                            done := make(chan struct{})
                                go func() {
                                    runner.DumpAST()
                                    close(done)
                                }()
                            <-done

                            // Cleaning up
                            err = runner.Finish()
                            if err != nil {
                                pterm.Error.Println(err.Error())
                                return err
                            }
                        }
                    }

                    // Wait for all sub-processes to finish
                    outWriter.Flush()

                    // Now we start the analysis of each AST file
                    pbAnalaysis1, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Main analysis")
                    spinnerAllExecution.UpdateTitle("Analyzing...")
                    spinnerAllExecution.Increment()
                    allResults := Analyzer.Start(pbAnalaysis1)

                    // Start aggregating results
                    spinnerAllExecution.UpdateTitle("Aggregating...")
                    aggregated := Analyzer.Aggregates(allResults)

                    spinnerAllExecution.UpdateTitle("")
                    spinnerAllExecution.Stop()
                    multi.Stop()

                    // Inform user
                    pterm.Success.Println("Finished")

                   pterm.DefaultTable.WithBoxed().WithHasHeader().WithData(pterm.TableData{
                   		{"Classes", "Methods", "AVG methods per class", "Min cyclomatic complexity", "Max cyclomatic complexity"},
                   		{strconv.Itoa(aggregated.NbClasses), strconv.Itoa(aggregated.NbMethods), fmt.Sprintf("%.2f", aggregated.AverageMethodsPerClass), strconv.Itoa(aggregated.MinCyclomaticComplexity), strconv.Itoa(aggregated.MaxCyclomaticComplexity),},
                   	}).Render()

                   	pterm.Println() // Blank line


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