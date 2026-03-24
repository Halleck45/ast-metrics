package mcp

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/mark3labs/mcp-go/mcp"
)

func getFileMetricsTool() mcp.Tool {
	return mcp.NewTool("get_file_metrics",
		mcp.WithDescription("Get detailed metrics for a specific file: LOC, complexity, maintainability, Halstead metrics, risk score, coupling, and per-class/method breakdown."),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path (relative to project root) to get metrics for")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get File Metrics",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	)
}

func handleGetFileMetrics(svc *AnalysisService) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError("Missing required parameter: path"), nil
		}

		agg, _, err := svc.Analyze(false)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Analysis failed: %v", err)), nil
		}

		// Find matching file
		for _, f := range agg.Combined.ConcernedFiles {
			if !matchesPath(f.Path, path) {
				continue
			}

			result := buildFileMetrics(f)

			// Add classes
			classes := engine.GetClassesInFile(f)
			var classMetrics []map[string]any
			for _, c := range classes {
				cm := map[string]any{
					"name": getClassName(c),
				}
				if c.Stmts != nil && c.Stmts.Analyze != nil {
					fillAnalyzeMetrics(cm, c.Stmts.Analyze)
				}
				classMetrics = append(classMetrics, cm)
			}
			if len(classMetrics) > 0 {
				result["classes"] = classMetrics
			}

			// Add functions
			functions := engine.GetFunctionsInFile(f)
			var funcMetrics []map[string]any
			for _, fn := range functions {
				fm := map[string]any{
					"name": getFuncName(fn),
				}
				if fn.Stmts != nil && fn.Stmts.Analyze != nil {
					fillAnalyzeMetrics(fm, fn.Stmts.Analyze)
				}
				funcMetrics = append(funcMetrics, fm)
			}
			if len(funcMetrics) > 0 {
				result["functions"] = funcMetrics
			}

			return safeToolResultJSON(result)
		}

		return mcp.NewToolResultError(fmt.Sprintf("File not found in analysis: %s", path)), nil
	}
}

func matchesPath(filePath, query string) bool {
	// Exact match or suffix match
	if filePath == query {
		return true
	}
	cleanQuery := filepath.Clean(query)
	cleanFile := filepath.Clean(filePath)
	return strings.HasSuffix(cleanFile, cleanQuery) || strings.HasSuffix(cleanFile, "/"+cleanQuery)
}
