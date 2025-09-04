package command

import (
	"bufio"
	"os"
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/storage"
	"github.com/pterm/pterm"
)

func TestAnalyzeCommand_Execute(t *testing.T) {
	t.Run("TestAnalyzeCommand_Execute", func(t *testing.T) {
		// Setup
		storage := storage.Default()
		storage.Purge()
		storage.Ensure()

		// HTML report
		tmpReportHtmlDir := t.TempDir()

		// Markdown report
		tmpReportMarkdownDir := t.TempDir() + "/report.md"

		// Sources
		sourcesDir1 := t.TempDir()
		// Add some files
		file1 := sourcesDir1 + "/test.php"
		os.Create(file1)
		sourcesToAnalyze := []string{sourcesDir1}

		// Configuration
		configuration := &configuration.Configuration{
			SourcesToAnalyzePath: sourcesToAnalyze,
			Reports: configuration.ConfigurationReport{
				Html:     tmpReportHtmlDir,
				Markdown: tmpReportMarkdownDir,
			},
			Storage: storage,
		}
		// Create a new AnalyzeCommand

		// create fake *bufio.Writer
		outWriter := bufio.NewWriter(os.Stdout)
		cmd := NewAnalyzeCommand(configuration, outWriter, nil, false)

		// Execute the command
		err := cmd.Execute()

		// Check if there was an error
		if err != nil {
			t.Errorf("AnalyzeCommand.Execute() = %s; want it to be nil", err.Error())
		}

		// fake pterm *pterm.SpinnerPrinter
		spinner := pterm.DefaultSpinner
		// Check if the analysis was done
		allResults := analyzer.Start(storage, &spinner)
		if allResults == nil {
			t.Errorf("analyzer.Start() = nil; want it to not be nil")
		}

		// Check HTML report has been created
		_, err = os.Stat(tmpReportHtmlDir + "/index.html")
		if err != nil {
			t.Errorf("os.Stat(tmpReportHtmlDir + \"/index.html\") = %s; want it to be nil", err.Error())
		}

		// Check markdown report file has been created
		_, err = os.Stat(tmpReportMarkdownDir)
		if err != nil {
			t.Errorf("os.Stat(tmpReportMarkdownDir) = %s; want it to be nil", err.Error())
		}
	})
}
