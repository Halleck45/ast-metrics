package mcp

import (
	"context"
	"fmt"
	"sort"

	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/mark3labs/mcp-go/mcp"
)

func listComponentsTool() mcp.Tool {
	return mcp.NewTool("list_components",
		mcp.WithDescription("List all components (packages, classes, files) found in the project. Use this to discover component names before calling get_coupling or get_dependencies."),
		mcp.WithBoolean("force_refresh", mcp.Description("Force re-analysis ignoring cache")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "List Components",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	)
}

func handleListComponents(svc *AnalysisService) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		// Graph nodes (packages/namespaces)
		var graphNodes []string
		if agg.Combined.Graph != nil {
			for id := range agg.Combined.Graph.Nodes {
				graphNodes = append(graphNodes, id)
			}
			sort.Strings(graphNodes)
		}

		// Classes
		var classes []map[string]string
		for _, f := range agg.Combined.ConcernedFiles {
			for _, c := range engine.GetClassesInFile(f) {
				classes = append(classes, map[string]string{
					"name": getClassName(c),
					"file": f.Path,
				})
			}
		}

		// Files with their language
		var files []map[string]string
		for _, f := range agg.Combined.ConcernedFiles {
			files = append(files, map[string]string{
				"path":     f.Path,
				"language": f.ProgrammingLanguage,
			})
		}

		return safeToolResultJSON(map[string]any{
			"graph_nodes": graphNodes,
			"classes":     classes,
			"files":       files,
		})
	}
}
