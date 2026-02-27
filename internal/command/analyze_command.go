package command

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	Activity "github.com/halleck45/ast-metrics/internal/analyzer/activity"
	requirement "github.com/halleck45/ast-metrics/internal/analyzer/requirement"
	"github.com/halleck45/ast-metrics/internal/cli"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	filefinder "github.com/halleck45/ast-metrics/internal/file"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/halleck45/ast-metrics/internal/report"
	"github.com/inancgumus/screen"
	"github.com/pterm/pterm"
	log "github.com/sirupsen/logrus"
)

type AnalyzeCommand struct {
	Configuration   *configuration.Configuration
	outWriter       *bufio.Writer
	runners         []engine.Engine
	isInteractive   bool
	moonSpinner     *pterm.SpinnerPrinter
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

	if v.alreadyExecuted {
		v.moonSpinner = nil
		v.outWriter.Flush()
	}

	if v.isInteractive && !v.alreadyExecuted {
		fmt.Print(cli.ScreenHeader("Analyzing"))
		fmt.Println()
		v.moonSpinner, _ = cli.NewMoonSpinner("Preparing analysis...")
	}

	if v.alreadyExecuted {
		v.moonSpinner = nil
	}

	// Parse source code into in-memory ASTs
	parsedFiles, err := v.ExecuteRunnerAnalysis(v.Configuration)
	if err != nil {
		return err
	}

	if v.moonSpinner != nil {
		v.outWriter.Flush()
	}

	// Now we start the analysis of each parsed file
	if v.moonSpinner != nil {
		v.moonSpinner.UpdateText("Analyzing source code...")
	}

	// Run global analysis on in-memory files
	allResults := analyzer.AnalyzeFiles(parsedFiles, nil)

	// Git analysis
	if v.moonSpinner != nil {
		v.moonSpinner.UpdateText("Analyzing git history...")
	}
	if v.gitSummaries == nil {
		gitAnalyzer := analyzer.NewGitAnalyzer()
		v.gitSummaries = gitAnalyzer.Start(allResults)
	}

	// Now compare with another branch (if needed)
	allResultsCloned := []*pb.File{}

	if v.Configuration.CompareWith != "" {

		if v.moonSpinner != nil {
			v.moonSpinner.UpdateText("Comparing with " + v.Configuration.CompareWith + "...")
		}

		// switch branches
		for _, gitSummary := range v.gitSummaries {
			err = gitSummary.GitRepository.Checkout(v.Configuration.CompareWith)
			if err != nil {
				return errors.New(`Cannot compare code with branch or commit "` + v.Configuration.CompareWith +
					`" for repository ` + gitSummary.GitRepository.Path)
			}
		}

		// execute analysis on the other branch (reset file discovery cache)
		clonedConfig := *v.Configuration
		clonedConfig.FileDiscovery = nil
		parsedCloned, err := v.ExecuteRunnerAnalysis(&clonedConfig)
		if err != nil {
			return err
		}

		// Run global analysis on the other branch
		allResultsCloned = analyzer.AnalyzeFiles(parsedCloned, nil)

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

	if v.moonSpinner != nil {
		v.moonSpinner.UpdateText("Aggregating results...")
	}
	projectAggregated := aggregator.Aggregates()

	// Evaluate requirements generating reports so templates can use results
	if v.Configuration.Requirements != nil {
		requirementsEvaluator := requirement.NewRequirementsEvaluator(*v.Configuration.Requirements)
		evaluation := requirementsEvaluator.Evaluate(allResults, requirement.ProjectAggregated{})
		projectAggregated.Evaluation = &evaluation
	}

	// Generate reports
	if v.moonSpinner != nil {
		v.moonSpinner.UpdateText("Generating reports...")
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
				cli.PrintError("Cannot generate report: " + err.Error())
				return err
			}
			generatedReports = append(generatedReports, reports...)
		}
	}

	if v.moonSpinner != nil {
		v.moonSpinner.Stop()
	}

	// Details errors
	if len(projectAggregated.ErroredFiles) > 0 {
		cli.PrintWarning(fmt.Sprintf("%d files could not be analyzed. Use the --verbose option to get details", len(projectAggregated.ErroredFiles)))
		if log.GetLevel() == log.DebugLevel {
			for _, file := range projectAggregated.ErroredFiles {
				cli.PrintError("File " + file.Path)
				for _, err := range file.Errors {
					cli.PrintError("    " + err)
				}
			}
		}
	}

	// Interactive: ask user what to do next
	if v.isInteractive && !v.alreadyExecuted {
		choice := cli.AskPostAnalysis(allResults, projectAggregated)
		switch choice {
		case cli.PostAnalysisOpenHTML:
			cli.GenerateAndOpenHTMLReport(allResults, projectAggregated)
			v.alreadyExecuted = true
			return nil
		case cli.PostAnalysisExplore:
			// Fall through to show ScreenHome
		case cli.PostAnalysisQuit:
			v.alreadyExecuted = true
			return nil
		}
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

	// Link to file watcher (in order to close it when app is closed)
	if v.FileWatcher != nil {
		v.currentPage.FileWatcher = v.FileWatcher
	}

	// Store state of the command
	v.alreadyExecuted = true

	// End screen (non-interactive reports summary)
	if !v.isInteractive {
		endScreen := cli.NewScreenEnd(v.isInteractive, allResults, projectAggregated, *v.Configuration, generatedReports)
		endScreen.Render()
	}

	return nil
}

func (v *AnalyzeCommand) ExecuteRunnerAnalysis(config *configuration.Configuration) ([]*pb.File, error) {
	// Precompute file discovery for all languages in a single directory walk
	if config.FileDiscovery == nil {
		discovery := &filefinder.FileDiscovery{}
		finder := filefinder.Finder{Configuration: *config}
		discovery.Precompute(finder, []string{".go", ".php", ".py", ".rs"})
		config.FileDiscovery = discovery
	}

	var allParsed []*pb.File

	for _, runner := range v.runners {

		runner.SetConfiguration(config)

		if !runner.IsRequired() {
			continue
		}

		runner.SetProgressbar(nil)

		if v.moonSpinner != nil {
			v.moonSpinner.UpdateText("Parsing source files...")
		}

		err := runner.Ensure()
		if err != nil {
			cli.PrintError(err.Error())
			return nil, err
		}

		// Parse files (in parallel within DumpAST)
		if v.moonSpinner != nil {
			v.moonSpinner.UpdateText("Building AST...")
		}

		parsed := runner.DumpAST()
		allParsed = append(allParsed, parsed...)

		// Cleaning up
		err = runner.Finish()
		if err != nil {
			cli.PrintError(err.Error())
			// pass
		}
	}

	return allParsed, nil
}
