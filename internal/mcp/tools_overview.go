package mcp

import (
	"context"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
)

func analyzeProjectTool() mcp.Tool {
	return mcp.NewTool("analyze_project",
		mcp.WithDescription("Analyze the project and return a high-level overview: language breakdown, file/class/method counts, average complexity, maintainability, coupling, top risky files, and suggestions."),
		mcp.WithBoolean("force_refresh", mcp.Description("Force re-analysis ignoring cache")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Analyze Project",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	)
}

func handleAnalyzeProject(svc *AnalysisService) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		forceRefresh := false
		if args := request.GetArguments(); args != nil {
			if v, ok := args["force_refresh"].(bool); ok {
				forceRefresh = v
			}
		}

		agg, _, err := svc.Analyze(forceRefresh)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Analysis failed: %v", err)), nil
		}

		combined := agg.Combined

		// Language breakdown
		langs := make(map[string]int)
		for lang, count := range combined.ProgrammingLanguages {
			langs[lang] = count
		}

		// Top risky files (up to 10)
		type riskyFile struct {
			Path  string  `json:"path"`
			Risk  float64 `json:"risk_score"`
		}
		var risky []riskyFile
		for _, f := range combined.ConcernedFiles {
			if f.Stmts != nil && f.Stmts.Analyze != nil && f.Stmts.Analyze.Risk != nil {
				score := f.Stmts.Analyze.Risk.Score
				if score > 0.3 {
					risky = append(risky, riskyFile{Path: f.Path, Risk: score})
				}
			}
		}
		sort.Slice(risky, func(i, j int) bool { return risky[i].Risk > risky[j].Risk })
		if len(risky) > 10 {
			risky = risky[:10]
		}

		// Suggestions (up to 5)
		var suggestions []map[string]string
		for i, s := range combined.Suggestions {
			if i >= 5 {
				break
			}
			suggestions = append(suggestions, map[string]string{
				"summary":  s.Summary,
				"location": s.Location,
				"why":      s.Why,
			})
		}

		result := map[string]any{
			"languages":   langs,
			"files":       combined.NbFiles,
			"classes":     combined.NbClasses,
			"functions":   combined.NbFunctions,
			"methods":     combined.NbMethods,
			"loc":         map[string]any{"sum": combined.Loc.Sum, "avg": combined.Loc.Avg},
			"cyclomatic_complexity": map[string]any{
				"avg": combined.CyclomaticComplexity.Avg,
				"max": combined.CyclomaticComplexity.Max,
				"sum": combined.CyclomaticComplexity.Sum,
			},
			"maintainability_index": map[string]any{
				"avg": combined.MaintainabilityIndex.Avg,
				"min": combined.MaintainabilityIndex.Min,
			},
			"coupling": map[string]any{
				"afferent_avg":  combined.AfferentCoupling.Avg,
				"efferent_avg":  combined.EfferentCoupling.Avg,
				"instability_avg": combined.Instability.Avg,
			},
			"bus_factor":      combined.BusFactor,
			"top_risky_files": risky,
			"suggestions":     suggestions,
			"errored_files":   len(agg.ErroredFiles),
		}

		return safeToolResultJSON(result)
	}
}
