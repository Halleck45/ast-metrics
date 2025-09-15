package command

import (
	"bufio"
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/halleck45/ast-metrics/internal/analyzer"
	requirement "github.com/halleck45/ast-metrics/internal/analyzer/requirement"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/pterm/pterm"
)

// LintCommand runs the analysis and prints only requirement violations (lint), grouped by file.
type LintCommand struct {
	Configuration *configuration.Configuration
	outWriter     *bufio.Writer
	runners       []engine.Engine
	verbose       bool
}

func (c *LintCommand) SetVerbose(v bool) { c.verbose = v }

func NewLintCommand(configuration *configuration.Configuration, outWriter *bufio.Writer, runners []engine.Engine) *LintCommand {
	return &LintCommand{
		Configuration: configuration,
		outWriter:     outWriter,
		runners:       runners,
		verbose:       false,
	}
}

func (c *LintCommand) Execute() error {
	// Prepare workdir
	c.Configuration.Storage.Purge()
	c.Configuration.Storage.Ensure()

	// Run engines to dump ASTs
	for _, runner := range c.runners {
		runner.SetConfiguration(c.Configuration)
		if !runner.IsRequired() {
			continue
		}
		if err := runner.Ensure(); err != nil {
			return err
		}
		done := make(chan struct{})
		go func() {
			runner.DumpAST()
			close(done)
		}()
		<-done
		if err := runner.Finish(); err != nil {
			return err
		}
	}

	// Global analysis (no UI/report)
	allResults := analyzer.Start(c.Configuration.Storage, nil)

	// Evaluate requirements
	if c.Configuration.Requirements == nil {
		pterm.Info.Println("No requirements configured. Nothing to lint.")
		return nil
	}
	reqEval := requirement.NewRequirementsEvaluator(*c.Configuration.Requirements)
	evaluation := reqEval.Evaluate(allResults, requirement.ProjectAggregated{})

	// Build a map[filePath][]outcomes directly from structured results
	grouped := map[string][]requirement.RuleOutcome{}
	ungrouped := []requirement.RuleOutcome{}
	for _, out := range evaluation.Errors {
		if out.File == "" {
			ungrouped = append(ungrouped, out)
			continue
		}
		grouped[out.File] = append(grouped[out.File], out)
	}

	// When verbose, also prepare successes grouped by file
	groupedOK := map[string][]requirement.RuleOutcome{}
	if c.verbose {
		for _, ok := range evaluation.Successes {
			if ok.File == "" {
				continue
			}
			groupedOK[ok.File] = append(groupedOK[ok.File], ok)
		}
	}

	// Pretty print lint by file
	totalHigh := 0
	totalMedium := 0
	totalLow := 0
	total := 0
	files := make([]string, 0, len(grouped))
	for f := range grouped {
		files = append(files, f)
	}
	// If verbose, include files that only have successes
	if c.verbose {
		for f := range groupedOK {
			found := false
			for _, existing := range files {
				if existing == f {
					found = true
					break
				}
			}
			if !found {
				files = append(files, f)
			}
		}
	}
	sort.Strings(files)

	for _, f := range files {
		underline := lipgloss.NewStyle().Underline(true).Bold(true)
		pterm.Println(underline.Render("File: " + f))

		// successes first if verbose
		if c.verbose {
			oks := groupedOK[f]
			sort.Slice(oks, func(i, j int) bool { return oks[i].Message < oks[j].Message })
			for _, s := range oks {
				pterm.Success.Println("  ✓ " + f + " — " + stripPathPrefix(s.Message, f))
			}
		}
		// sort messages for deterministic output
		msgs := grouped[f]
		sort.Slice(msgs, func(i, j int) bool { return msgs[i].Message < msgs[j].Message })
		for _, m := range msgs {
			badge := ""
			switch m.Severity {
			case requirement.SeverityHigh:
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
				badge = style.Render("HIGH    ")
				totalHigh++
			case requirement.SeverityMedium:
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
				badge = style.Render("MEDIUM  ")
				totalMedium++
			case requirement.SeverityLow:
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("#008000")).Bold(true)
				badge = style.Render("LOW     ")
				totalLow++
			}

			greyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

			ruleStyled := greyStyle.Render(" #" + m.Rule + "")
			content := " • " + badge + stripPathPrefix(m.Message, f) + ruleStyled
			pterm.Println(content)

			total++
		}
		pterm.Println()
	}

	if len(ungrouped) > 0 {
		pterm.Println("Other")
		sort.Slice(ungrouped, func(i, j int) bool { return ungrouped[i].Message < ungrouped[j].Message })
		for _, m := range ungrouped {
			badge := ""
			switch m.Severity {
			case requirement.SeverityHigh:
				badge = "[HIGH] "
				pterm.Error.Println("  • " + badge + m.Message)
			case requirement.SeverityMedium:
				badge = "[MED] "
				pterm.Warning.Println("  • " + badge + m.Message)
			case requirement.SeverityLow:
				badge = "[LOW] "
				pterm.Warning.Println("  • " + badge + m.Message)
			}

			total++
		}
		pterm.Println()
	}

	// Summary and exit code
	if total == 0 {
		return nil
	}

	// return new Error
	return fmt.Errorf("%d lint issue(s) found (%d high, %d medium, %d low)", total, totalHigh, totalMedium, totalLow)
}

// extractPath tries to match a File.Path from analysis results inside the message string
func extractPath(msg string, files []*pb.File) string {
	for _, f := range files {
		if f.Path != "" && contains(msg, f.Path) {
			return f.Path
		}
	}
	return ""
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	// simple implementation to avoid importing strings for now
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// stripPathPrefix removes the file path info from a message, preserving the rule description.
// It specifically strips patterns like " in file <path>".
func stripPathPrefix(msg, path string) string {
	// Try the common pattern first
	pattern := " in file " + path
	if i := indexOf(msg, pattern); i >= 0 {
		res := msg[:i] + msg[i+len(pattern):]
		return res
	}
	// Fallback: remove just the path, keeping surrounding text
	if i := indexOf(msg, path); i >= 0 {
		res := msg[:i] + msg[i+len(path):]
		return res
	}
	return msg
}
