package command

import (
	"bufio"
	"errors"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	Activity "github.com/halleck45/ast-metrics/internal/analyzer/activity"
	"github.com/halleck45/ast-metrics/internal/cli"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/halleck45/ast-metrics/internal/report"
	"github.com/halleck45/ast-metrics/internal/storage"
	"github.com/inancgumus/screen"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
)

type AnalyzeCommand struct {
	Configuration   *configuration.Configuration
	outWriter       *bufio.Writer
	runners         []engine.Engine
	isInteractive   bool
	spinner         *pterm.ProgressbarPrinter
	multi           *pterm.MultiPrinter
	alreadyExecuted bool
	currentPage     *cli.ScreenHome
	FileWatcher     *fsnotify.Watcher
	gitSummaries    []analyzer.ResultOfGitAnalysis
}

func NewAnalyzeCommand(configuration *configuration.Configuration, outWriter *bufio.Writer, runners []engine.Engine, isInteractive bool) *AnalyzeCommand {
	return &AnalyzeCommand{
		Configuration:   configuration,
		outWriter:       outWriter,
		runners:         runners,
		isInteractive:   isInteractive,
		alreadyExecuted: false,
	}
}

func (v *AnalyzeCommand) Execute() error {

	// Prepare workdir
	v.Configuration.Storage.Purge()
	v.Configuration.Storage.Ensure()

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

	// Convert source code to ASTs (each source code is converted to a binary protobuf file)
	err := v.ExecuteRunnerAnalysis(v.Configuration)
	if err != nil {
		return err
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

	// Run global analysis
	allResults := analyzer.Start(v.Configuration.Storage, progressBarAnalysis)

	if progressBarAnalysis != nil {
		progressBarAnalysis.Stop()
	}

	// Git analysis
	if v.spinner != nil {
		v.spinner.UpdateTitle("Git analysis...")
		v.spinner.Increment()
	}
	if v.gitSummaries == nil {
		gitAnalyzer := analyzer.NewGitAnalyzer()
		v.gitSummaries = gitAnalyzer.Start(allResults)
	}
	if progressBarAnalysis != nil {
		progressBarAnalysis.WithRemoveWhenDone(true)
		progressBarAnalysis.Stop()
		v.outWriter.Flush()
	}

	// Now compare with another branch (if needed)
	clonedConfiguration := v.Configuration
	allResultsCloned := []*pb.File{}

	if v.Configuration.CompareWith != "" {

		if v.spinner != nil {
			v.spinner.UpdateTitle("Comparing with " + v.Configuration.CompareWith)
			v.spinner.Increment()
		}

		// switch branches
		for _, gitSummary := range v.gitSummaries {
			err = gitSummary.GitRepository.Checkout(v.Configuration.CompareWith)
			if err != nil {
				return errors.New(`Cannot compare code with branch or commit "` + v.Configuration.CompareWith +
					`" for repository ` + gitSummary.GitRepository.Path)
			}
		}

		// create another workdir
		clonedConfiguration.Storage = storage.NewWithName("compare")
		clonedConfiguration.Storage.Purge()
		clonedConfiguration.Storage.Ensure()

		// execute analysis on the other branch
		err := v.ExecuteRunnerAnalysis(clonedConfiguration)
		if err != nil {
			return err
		}

		// Run global analysis on the other branch
		allResultsCloned = analyzer.Start(clonedConfiguration.Storage, progressBarAnalysis)

		// switch back to the original branch
		for _, gitSummary := range v.gitSummaries {
			err = gitSummary.GitRepository.RestoreFirstBranch()
			if err != nil {
				log.Error("Cannot checkout back to branch " + gitSummary.GitRepository.InitialBranch + " for " + gitSummary.GitRepository.Path)
			}
		}
	}

	// Start aggregating results
	aggregator := analyzer.NewAggregator(allResults, v.gitSummaries)
	aggregator.WithAggregateAnalyzer(Activity.NewBusFactor())
	if v.Configuration.CompareWith != "" {
		aggregator.WithComparaison(allResultsCloned, v.Configuration.CompareWith)
	}

	if v.spinner != nil {
		v.spinner.UpdateTitle("Aggregating...")
	}
	projectAggregated := aggregator.Aggregates()

	// Generate reports
	if v.spinner != nil {
		v.spinner.UpdateTitle("Generating reports...")
		v.spinner.Increment()
	}

	// Factory reporters
	reportersFactory := report.ReportersFactory{
		Configuration: v.Configuration,
	}
	reporters := reportersFactory.Factory(v.Configuration)

	generatedReports := []report.GeneratedReport{}
	if v.Configuration.Reports.HasReports() {
		// Generate reports
		for _, reporter := range reporters {
			reports, err := reporter.Generate(allResults, projectAggregated)
			if err != nil {
				pterm.Error.Println("Cannot generate report: " + err.Error())
				return err
			}
			generatedReports = append(generatedReports, reports...)
		}
	}

	// Evaluate requirements
	shouldFail := false
	if v.Configuration.Requirements != nil {
		requirementsEvaluator := analyzer.NewRequirementsEvaluator(*v.Configuration.Requirements)
		evaluation := requirementsEvaluator.Evaluate(allResults, projectAggregated)
		projectAggregated.Evaluation = &evaluation

		if evaluation.Succeeded {
			pterm.Success.Println("Requirements are met")
		} else {
			pterm.Error.Printf("Requirements are not met. Found %d violation(s)\n", len(evaluation.Errors))
			for _, err := range evaluation.Errors {
				pterm.Error.Println("    " + err)
			}

			shouldFail = v.Configuration.Requirements.FailOnError
		}
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
		v.currentPage = cli.NewScreenHome(v.isInteractive, allResults, projectAggregated)
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

	// End screen
	screen := cli.NewScreenEnd(v.isInteractive, allResults, projectAggregated, *v.Configuration, generatedReports)
	screen.Render()

	if shouldFail {
		os.Exit(1)
	}

	return nil
}

func (v *AnalyzeCommand) ExecuteRunnerAnalysis(config *configuration.Configuration) error {
	for _, runner := range v.runners {

		runner.SetConfiguration(config)

		if !runner.IsRequired() {
			continue
		}

		var progressBarSpecificForengine *pterm.ProgressbarPrinter = nil
		if v.spinner != nil {
			progressBarSpecificForengine, _ := pterm.DefaultSpinner.WithWriter(v.multi.NewWriter()).Start("...")
			progressBarSpecificForengine.RemoveWhenDone = true
			runner.SetProgressbar(progressBarSpecificForengine)
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
		if v.isInteractive && progressBarSpecificForengine != nil {
			progressBarSpecificForengine.Stop()
		}
		if err != nil {
			pterm.Error.Println(err.Error())
			// pass
		}
	}

	return nil
}
