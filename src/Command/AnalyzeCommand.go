package Command

import (
	"bufio"

	"github.com/halleck45/ast-metrics/src/Analyzer"
	"github.com/halleck45/ast-metrics/src/Cli"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
)

type AnalyzeCommand struct {
	configuration *Configuration.Configuration
	outWriter     *bufio.Writer
	runners       []Engine.Engine
	isInteractive bool
}

func NewAnalyzeCommand(configuration *Configuration.Configuration, outWriter *bufio.Writer, runners []Engine.Engine, isInteractive bool) *AnalyzeCommand {
	return &AnalyzeCommand{
		configuration: configuration,
		outWriter:     outWriter,
		runners:       runners,
		isInteractive: isInteractive,
	}
}

func (v *AnalyzeCommand) Execute() error {

	// Prepare workdir
	Storage.Purge()
	Storage.Ensure()

	// Prepare progress bars
	multi := pterm.DefaultMultiPrinter.WithWriter(v.outWriter)
	spinnerAllExecution, _ := pterm.DefaultProgressbar.WithTotal(3).WithWriter(multi.NewWriter()).WithTitle("Analyzing").Start()
	spinnerAllExecution.RemoveWhenDone = true
	defer spinnerAllExecution.Stop()

	// Start progress bars
	multi.Start()

	for _, runner := range v.runners {

		runner.SetConfiguration(v.configuration)

		if runner.IsRequired() {

			progressBarSpecificForEngine, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("...")
			progressBarSpecificForEngine.RemoveWhenDone = true
			runner.SetProgressbar(progressBarSpecificForEngine)

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
			progressBarSpecificForEngine.Stop()
			if err != nil {
				pterm.Error.Println(err.Error())
				// pass
			}
		}
	}

	// Wait for all sub-processes to finish
	v.outWriter.Flush()

	// Now we start the analysis of each AST file
	progressBarAnalysis, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("Main analysis")
	progressBarAnalysis.RemoveWhenDone = true
	spinnerAllExecution.UpdateTitle("Analyzing...")
	spinnerAllExecution.Increment()
	allResults := Analyzer.Start(progressBarAnalysis)
	progressBarAnalysis.Stop()

	// Start aggregating results
	aggregator := Analyzer.NewAggregator(allResults)
	spinnerAllExecution.UpdateTitle("Aggregating...")
	projectAggregated := aggregator.Aggregates()

	spinnerAllExecution.UpdateTitle("")
	spinnerAllExecution.Stop()
	multi.Stop()

	// Display results
	renderer := Cli.NewScreenHome(v.isInteractive, allResults, projectAggregated)
	renderer.Render()

	return nil
}
