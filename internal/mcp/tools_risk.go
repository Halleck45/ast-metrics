package mcp

import (
	"context"
	"fmt"
	"sort"

	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/mark3labs/mcp-go/mcp"
)

func findRiskyCodeTool() mcp.Tool {
	return mcp.NewTool("find_risky_code",
		mcp.WithDescription("Find files and classes with the highest risk scores. Risk combines complexity, maintainability, coupling, and change frequency."),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default: 20)")),
		mcp.WithNumber("min_risk", mcp.Description("Minimum risk score threshold 0-1 (default: 0.1)")),
		mcp.WithBoolean("force_refresh", mcp.Description("Force re-analysis ignoring cache")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Find Risky Code",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	)
}

func handleFindRiskyCode(svc *AnalysisService) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		limit := 20
		minRisk := 0.1
		forceRefresh := false
		if args != nil {
			if v, ok := args["limit"].(float64); ok {
				limit = int(v)
			}
			if v, ok := args["min_risk"].(float64); ok {
				minRisk = v
			}
			if v, ok := args["force_refresh"].(bool); ok {
				forceRefresh = v
			}
		}

		agg, _, err := svc.Analyze(forceRefresh)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Analysis failed: %v", err)), nil
		}

		type riskyItem struct {
			Path                 string  `json:"path"`
			Name                 string  `json:"name,omitempty"`
			Type                 string  `json:"type"` // "file" or "class"
			RiskScore            float64 `json:"risk_score"`
			CyclomaticComplexity *int32  `json:"cyclomatic_complexity,omitempty"`
			MaintainabilityIndex *float64 `json:"maintainability_index,omitempty"`
		}

		var items []riskyItem

		for _, f := range agg.Combined.ConcernedFiles {
			if f.Stmts == nil || f.Stmts.Analyze == nil || f.Stmts.Analyze.Risk == nil {
				continue
			}
			score := f.Stmts.Analyze.Risk.Score
			if score < minRisk {
				continue
			}

			item := riskyItem{
				Path:      f.Path,
				Type:      "file",
				RiskScore: score,
			}
			if f.Stmts.Analyze.Complexity != nil {
				item.CyclomaticComplexity = f.Stmts.Analyze.Complexity.Cyclomatic
			}
			if f.Stmts.Analyze.Maintainability != nil {
				item.MaintainabilityIndex = f.Stmts.Analyze.Maintainability.MaintainabilityIndex
			}
			items = append(items, item)

			// Also check classes within the file
			for _, c := range engine.GetClassesInFile(f) {
				if c.Stmts == nil || c.Stmts.Analyze == nil || c.Stmts.Analyze.Risk == nil {
					continue
				}
				cScore := c.Stmts.Analyze.Risk.Score
				if cScore < minRisk {
					continue
				}
				cItem := riskyItem{
					Path:      f.Path,
					Name:      getClassName(c),
					Type:      "class",
					RiskScore: cScore,
				}
				if c.Stmts.Analyze.Complexity != nil {
					cItem.CyclomaticComplexity = c.Stmts.Analyze.Complexity.Cyclomatic
				}
				if c.Stmts.Analyze.Maintainability != nil {
					cItem.MaintainabilityIndex = c.Stmts.Analyze.Maintainability.MaintainabilityIndex
				}
				items = append(items, cItem)
			}
		}

		sort.Slice(items, func(i, j int) bool { return items[i].RiskScore > items[j].RiskScore })
		if len(items) > limit {
			items = items[:limit]
		}

		return safeToolResultJSON(map[string]any{
			"risky_code": items,
			"total":      len(items),
		})
	}
}

func findComplexCodeTool() mcp.Tool {
	return mcp.NewTool("find_complex_code",
		mcp.WithDescription("Find functions and classes with high cyclomatic complexity or low maintainability index."),
		mcp.WithNumber("max_cyclomatic", mcp.Description("Complexity threshold - return items above this (default: 10)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default: 20)")),
		mcp.WithBoolean("force_refresh", mcp.Description("Force re-analysis ignoring cache")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Find Complex Code",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	)
}

func handleFindComplexCode(svc *AnalysisService) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()
		maxCyclomatic := int32(10)
		limit := 20
		forceRefresh := false
		if args != nil {
			if v, ok := args["max_cyclomatic"].(float64); ok {
				maxCyclomatic = int32(v)
			}
			if v, ok := args["limit"].(float64); ok {
				limit = int(v)
			}
			if v, ok := args["force_refresh"].(bool); ok {
				forceRefresh = v
			}
		}

		agg, _, err := svc.Analyze(forceRefresh)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Analysis failed: %v", err)), nil
		}

		type complexItem struct {
			Path                 string   `json:"path"`
			Name                 string   `json:"name"`
			Type                 string   `json:"type"` // "function", "class", "method"
			CyclomaticComplexity int32    `json:"cyclomatic_complexity"`
			MaintainabilityIndex *float64 `json:"maintainability_index,omitempty"`
			Loc                  *int32   `json:"loc,omitempty"`
		}

		var items []complexItem

		for _, f := range agg.Combined.ConcernedFiles {
			// Check classes
			for _, c := range engine.GetClassesInFile(f) {
				if c.Stmts == nil || c.Stmts.Analyze == nil || c.Stmts.Analyze.Complexity == nil || c.Stmts.Analyze.Complexity.Cyclomatic == nil {
					continue
				}
				cc := *c.Stmts.Analyze.Complexity.Cyclomatic
				if cc >= maxCyclomatic {
					item := complexItem{
						Path:                 f.Path,
						Name:                 getClassName(c),
						Type:                 "class",
						CyclomaticComplexity: cc,
					}
					if c.Stmts.Analyze.Maintainability != nil {
						item.MaintainabilityIndex = c.Stmts.Analyze.Maintainability.MaintainabilityIndex
					}
					if c.Stmts.Analyze.Volume != nil {
						item.Loc = c.Stmts.Analyze.Volume.Loc
					}
					items = append(items, item)
				}
			}

			// Check functions
			for _, fn := range engine.GetFunctionsInFile(f) {
				if fn.Stmts == nil || fn.Stmts.Analyze == nil || fn.Stmts.Analyze.Complexity == nil || fn.Stmts.Analyze.Complexity.Cyclomatic == nil {
					continue
				}
				cc := *fn.Stmts.Analyze.Complexity.Cyclomatic
				if cc >= maxCyclomatic {
					item := complexItem{
						Path:                 f.Path,
						Name:                 getFuncName(fn),
						Type:                 "function",
						CyclomaticComplexity: cc,
					}
					if fn.Stmts.Analyze.Maintainability != nil {
						item.MaintainabilityIndex = fn.Stmts.Analyze.Maintainability.MaintainabilityIndex
					}
					if fn.Stmts.Analyze.Volume != nil {
						item.Loc = fn.Stmts.Analyze.Volume.Loc
					}
					items = append(items, item)
				}
			}
		}

		sort.Slice(items, func(i, j int) bool { return items[i].CyclomaticComplexity > items[j].CyclomaticComplexity })
		if len(items) > limit {
			items = items[:limit]
		}

		return safeToolResultJSON(map[string]any{
			"complex_code": items,
			"total":        len(items),
		})
	}
}
