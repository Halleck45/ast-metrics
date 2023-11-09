package Command

import (
    "bufio"
    "os"
    "github.com/pterm/pterm"
    "github.com/halleck45/ast-metrics/src/Storage"
    "github.com/halleck45/ast-metrics/src/Engine"
    "github.com/halleck45/ast-metrics/src/Configuration"
    "github.com/halleck45/ast-metrics/src/Analyzer"
    "github.com/halleck45/ast-metrics/src/Driver"
    "github.com/halleck45/ast-metrics/src/Cli"
)

type AnalyzeCommand struct {
    path string
    driver Driver.Driver
    outWriter *bufio.Writer
    runners []Engine.Engine
    isInteractive bool
}

func NewAnalyzeCommand(configuration *Configuration.Configuration, outWriter *bufio.Writer, runners []Engine.Engine, isInteractive bool) *AnalyzeCommand {
    return &AnalyzeCommand{
        path: configuration.SourcesToAnalyzePath[0], // @todo: handle multiple paths
        driver: configuration.Driver,
        outWriter: outWriter,
        runners: runners,
        isInteractive: isInteractive,
    }
}

func (v *AnalyzeCommand) Execute() error {

    // ensure path exists
    if _, err := os.Stat(v.path); err != nil {
       pterm.Error.Println("Path '" + v.    path + "' does not exist or is not readable")
       return err
    }

    // Prepare workdir
    Storage.Ensure()

    // Prepare progress bars
    multi := pterm.DefaultMultiPrinter.WithWriter(v.outWriter)
    spinnerAllExecution, _ := pterm.DefaultProgressbar.WithTotal(3).WithWriter(multi.NewWriter()).WithTitle("Analyzing").Start()
    spinnerAllExecution.RemoveWhenDone = true

    // Start progress bars
    multi.Start()

    for _, runner := range v.runners {

        runner.SetDriver(v.driver)
        progressBarSpecificForEngine, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Checking PHP Engine")
        runner.SetProgressbar(progressBarSpecificForEngine)
        runner.SetSourcesToAnalyzePath(v.path)

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
    v.outWriter.Flush()

    // Now we start the analysis of each AST file
    progressBarAnalysis, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Main analysis")
    spinnerAllExecution.UpdateTitle("Analyzing...")
    spinnerAllExecution.Increment()
    allResults := Analyzer.Start(progressBarAnalysis)

    // Start aggregating results
    spinnerAllExecution.UpdateTitle("Aggregating...")
    aggregated := Analyzer.Aggregates(allResults)

    spinnerAllExecution.UpdateTitle("")
    spinnerAllExecution.Stop()
    multi.Stop()

    // Dislpay results
    // @todo: move this to a renderer and use a loop of renderers
    Cli.AggregationSummary(aggregated)

    renderer := Cli.NewRendererTableClass(v.isInteractive)
    renderer.Render(allResults, aggregated)

    return nil
}