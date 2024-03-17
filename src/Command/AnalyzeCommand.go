package Command

import (
	"bufio"

	"github.com/halleck45/ast-metrics/src/Analyzer"
	Activity "github.com/halleck45/ast-metrics/src/Analyzer/Activity"
	"github.com/halleck45/ast-metrics/src/Cli"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	Report "github.com/halleck45/ast-metrics/src/Report/Html"
	Markdown "github.com/halleck45/ast-metrics/src/Report/Markdown"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/pterm/pterm"
)

type AnalyzeCommand struct {
	configuration *Configuration.Configuration
	outWriter     *bufio.Writer
	runners       []Engine.Engine
	isInteractive bool
	spinner       *pterm.ProgressbarPrinter
	multi         *pterm.MultiPrinter
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

	if v.isInteractive {
		// Prepare progress bars
		v.multi = pterm.DefaultMultiPrinter.WithWriter(v.outWriter)
		v.spinner, _ = pterm.DefaultProgressbar.WithTotal(7).WithWriter(v.multi.NewWriter()).WithTitle("Analyzing").Start()
		v.spinner.RemoveWhenDone = true
		defer v.spinner.Stop()

		// Start progress bars
		v.multi.Start()
	}

	for _, runner := range v.runners {

		runner.SetConfiguration(v.configuration)

		if runner.IsRequired() {

			progressBarSpecificForEngine, _ := pterm.DefaultSpinner.WithWriter(v.multi.NewWriter()).Start("...")
			progressBarSpecificForEngine.RemoveWhenDone = true
			runner.SetProgressbar(progressBarSpecificForEngine)

			if v.spinner != nil {
				v.spinner.Increment()
			}

			err := runner.Ensure()
			if err != nil {
				pterm.Error.Println(err.Error())
				return err
			}

			// Dump ASTs (in parallel)
			if v.spinner != nil {
				v.spinner.UpdateTitle("Dumping AST code...")
				v.spinner.Increment()
			}

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

	v.outWriter.Flush()

	// Now we start the analysis of each AST file
	var progressBarAnalysis *pterm.SpinnerPrinter
	progressBarAnalysis = nil
	if v.spinner != nil {
		progressBarAnalysis, _ := pterm.DefaultSpinner.WithWriter(v.multi.NewWriter()).Start("Main analysis")
		progressBarAnalysis.RemoveWhenDone = true
		v.spinner.UpdateTitle("Analyzing...")
		v.spinner.Increment()
	}

	allResults := Analyzer.Start(progressBarAnalysis)
	if progressBarAnalysis != nil {
		progressBarAnalysis.Stop()
	}

	// Git analysis
	if v.spinner != nil {
		v.spinner.UpdateTitle("Git analysis...")
		v.spinner.Increment()
	}
	gitAnalyzer := Analyzer.NewGitAnalyzer()
	gitAnalyzer.Start(allResults)
	if progressBarAnalysis != nil {
		progressBarAnalysis.Stop()
		v.outWriter.Flush()
	}

	// Start aggregating results
	aggregator := Analyzer.NewAggregator(allResults)
	aggregator.WithAggregateAnalyzer(Activity.NewBusFactor())
	if v.spinner != nil {
		v.spinner.UpdateTitle("Aggregating...")
		//v.spinner.Increment()
	}
	projectAggregated := aggregator.Aggregates()

	// Generate reports
	if v.spinner != nil {
		v.spinner.UpdateTitle("Generating reports...")
		v.spinner.Increment()
	}

	// report: html
	htmlReportGenerator := Report.NewHtmlReportGenerator(v.configuration.HtmlReportPath)
	err := htmlReportGenerator.Generate(allResults, projectAggregated)
	if err != nil {
		pterm.Error.Println(err.Error())
		return err
	}
	// report: markdown
	markdownReportGenerator := Markdown.NewMarkdownReportGenerator(v.configuration.MarkdownReportPath)
	err = markdownReportGenerator.Generate(allResults, projectAggregated)
	if err != nil {
		pterm.Error.Println(err.Error())
	}

	if v.spinner != nil {
		v.spinner.UpdateTitle("")
		v.spinner.Stop()
		v.multi.Stop()
	}

	// Display results
	renderer := Cli.NewScreenHome(v.isInteractive, allResults, projectAggregated)
	renderer.Render()

	return nil
}
