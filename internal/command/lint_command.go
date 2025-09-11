package command

import (
	"bufio"
	"fmt"
	"os"
	"sort"

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

	// Build a map[filePath][]messages from evaluation.Errors when possible.
	// Our evaluation returns only strings; try to find file path presence inside messages.
	grouped := map[string][]string{}
	ungrouped := []string{}
	for _, msg := range evaluation.Errors {
		path := extractPath(msg, allResults)
		if path == "" {
			ungrouped = append(ungrouped, msg)
			continue
		}
		grouped[path] = append(grouped[path], msg)
	}

	// When verbose, also prepare successes grouped by file
	groupedOK := map[string][]string{}
	if c.verbose {
		for _, ok := range evaluation.Successes {
			p := extractPath(ok, allResults)
			if p == "" {
				continue
			}
			groupedOK[p] = append(groupedOK[p], ok)
		}
	}

	// Pretty print lint by file
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
				if existing == f { found = true; break }
			}
			if !found { files = append(files, f) }
		}
	}
	sort.Strings(files)

	for _, f := range files {
		pterm.Println("File:", f)
		// successes first if verbose
		if c.verbose {
			oks := groupedOK[f]
			sort.Strings(oks)
			for _, s := range oks {
				pterm.Success.Println("  ✓ " + f + " — " + stripPathPrefix(s, f))
			}
		}
		// sort messages for deterministic output
		msgs := grouped[f]
		sort.Strings(msgs)
		for _, m := range msgs {
			pterm.Error.Println("  • " + stripPathPrefix(m, f))
			total++
		}
		pterm.Println()
	}

	if len(ungrouped) > 0 {
 	pterm.Println("Other")
		sort.Strings(ungrouped)
		for _, m := range ungrouped {
			pterm.Error.Println("  • " + m)
			total++
		}
		pterm.Println()
	}

	// Summary and exit code
	if total == 0 {
		pterm.Success.Println("No lint issues found. Requirements are met.")
		return nil
	}

	pterm.Error.Printf("%d lint issue(s) found\n", total)
	if c.Configuration.Requirements.FailOnError {
		os.Exit(1)
	}
	return fmt.Errorf("%d lint issue(s) found", total)
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
