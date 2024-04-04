package Command

import (
	"bufio"

	"github.com/fsnotify/fsnotify"
	"github.com/halleck45/ast-metrics/src/Analyzer"
	Activity "github.com/halleck45/ast-metrics/src/Analyzer/Activity"
	"github.com/halleck45/ast-metrics/src/Cli"
	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/halleck45/ast-metrics/src/Engine"
	Report "github.com/halleck45/ast-metrics/src/Report/Html"
	Markdown "github.com/halleck45/ast-metrics/src/Report/Markdown"
	"github.com/halleck45/ast-metrics/src/Storage"
	"github.com/inancgumus/screen"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
)

type AnalyzeCommand struct {
	configuration   *Configuration.Configuration
	outWriter       *bufio.Writer
	runners         []Engine.Engine
	isInteractive   bool
	spinner         *pterm.ProgressbarPrinter
	multi           *pterm.MultiPrinter
	alreadyExecuted bool
	currentPage     *Cli.ScreenHome
	FileWatcher     *fsnotify.Watcher
	gitSummaries    []Analyzer.ResultOfGitAnalysis
}

func NewAnalyzeCommand(configuration *Configuration.Configuration, outWriter *bufio.Writer, runners []Engine.Engine, isInteractive bool) *AnalyzeCommand {
	return &AnalyzeCommand{
		configuration:   configuration,
		outWriter:       outWriter,
		runners:         runners,
		isInteractive:   isInteractive,
		alreadyExecuted: false,
	}
}

func (v *AnalyzeCommand) Execute() error {

	// Prepare workdir
	Storage.Purge()
	Storage.Ensure()

	if v.alreadyExecuted {
		// On refresh
		//v.spinner.Stop()
		v.spinner = nil
		// clean
		v.outWriter.Flush()
	}

	if v.isInteractive && !v.alreadyExecuted {
		// Prepare progress bars
		v.multi = pterm.DefaultMultiPrinter.WithWriter(v.outWriter)
		v.spinner, _ = pterm.DefaultProgressbar.WithTotal(7).WithWriter(v.multi.NewWriter()).WithTitle("Analyzing").Start()
		v.spinner.RemoveWhenDone = true
		defer v.spinner.Stop()

		// Start progress bars
		v.multi.Start()
	}

	if v.alreadyExecuted {
		v.spinner = nil
	}

	for _, runner := range v.runners {

		runner.SetConfiguration(v.configuration)

		if runner.IsRequired() {

			var progressBarSpecificForEngine *pterm.ProgressbarPrinter = nil
			if v.spinner != nil {
				progressBarSpecificForEngine, _ := pterm.DefaultSpinner.WithWriter(v.multi.NewWriter()).Start("...")
				progressBarSpecificForEngine.RemoveWhenDone = true
				runner.SetProgressbar(progressBarSpecificForEngine)
			}

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
			if v.isInteractive && progressBarSpecificForEngine != nil {
				progressBarSpecificForEngine.Stop()
			}
			if err != nil {
				pterm.Error.Println(err.Error())
				// pass
			}
		}
	}

	if v.spinner != nil {
		v.outWriter.Flush()
	}

	// Now we start the analysis of each AST file
	var progressBarAnalysis *pterm.SpinnerPrinter = nil
	if v.spinner != nil {
		progressBarAnalysis, _ = pterm.DefaultSpinner.WithWriter(v.multi.NewWriter()).Start("Main analysis")
		progressBarAnalysis.RemoveWhenDone = true
		v.spinner.UpdateTitle("Analyzing...")
		v.spinner.Increment()
	}

	workdir := Storage.Path()
	allResults := Analyzer.Start(workdir, progressBarAnalysis)
	if progressBarAnalysis != nil {
		progressBarAnalysis.Stop()
	}

	// Git analysis
	if v.spinner != nil {
		v.spinner.UpdateTitle("Git analysis...")
		v.spinner.Increment()
	}
	if v.gitSummaries == nil {
		gitAnalyzer := Analyzer.NewGitAnalyzer()
		v.gitSummaries = gitAnalyzer.Start(allResults)
	}
	if progressBarAnalysis != nil {
		progressBarAnalysis.WithRemoveWhenDone(true)
		progressBarAnalysis.Stop()
		v.outWriter.Flush()
	}

	// Start aggregating results
	aggregator := Analyzer.NewAggregator(allResults, v.gitSummaries)
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
	htmlReportGenerator := Report.NewHtmlReportGenerator(v.configuration.Reports.Html)
	err := htmlReportGenerator.Generate(allResults, projectAggregated)
	if err != nil {
		pterm.Error.Println("Cannot generate html report: " + err.Error())
		return err
	}
	// report: markdown
	markdownReportGenerator := Markdown.NewMarkdownReportGenerator(v.configuration.Reports.Markdown)
	err = markdownReportGenerator.Generate(allResults, projectAggregated)
	if err != nil {
		pterm.Error.Println("Cannot generate markdown report: " + err.Error())
	}

	if v.spinner != nil {
		v.spinner.UpdateTitle("")
		v.spinner.Stop()
		v.multi.Stop()
	}

	// Display results
	if v.currentPage == nil {
		if v.isInteractive {
			screen.Clear()
			screen.MoveTopLeft()
		}
		v.currentPage = Cli.NewScreenHome(v.isInteractive, allResults, projectAggregated)
		v.currentPage.Render()
	} else {
		screen.MoveTopLeft()
		v.currentPage.Reset(allResults, projectAggregated)
	}

	// Details errors
	if len(projectAggregated.ErroredFiles) > 0 {
		pterm.Info.Printf("%d files could not be analyzed. Use the --verbose option to get details\n", len(projectAggregated.ErroredFiles))
		if log.GetLevel() == log.DebugLevel {
			for _, file := range projectAggregated.ErroredFiles {
				pterm.Error.Println("File " + file.Path)
				for _, err := range file.Errors {
					pterm.Error.Println("    " + err)
				}
			}
		}
	}

	// Link to file wartcher (in order to close it when app is closed)
	if v.FileWatcher != nil {
		v.currentPage.FileWatcher = v.FileWatcher
	}

	// Store state of the command
	v.alreadyExecuted = true

	return nil
}
