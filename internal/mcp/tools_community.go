package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func getCommunitiesTool() mcp.Tool {
	return mcp.NewTool("get_communities",
		mcp.WithDescription("Get architectural community analysis: groups of related modules detected via dependency graph clustering. Shows community sizes, purity, coupling ratios, betweenness, bus factor, and inter-community edges."),
		mcp.WithBoolean("force_refresh", mcp.Description("Force re-analysis ignoring cache")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			Title:        "Get Communities",
			ReadOnlyHint: mcp.ToBoolPtr(true),
		}),
	)
}

func handleGetCommunities(svc *AnalysisService) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		cm := agg.Combined.Community
		if cm == nil {
			return mcp.NewToolResultError("No community analysis available (dependency graph may be too small)"), nil
		}

		// Build per-community details
		type communityDetail struct {
			ID              string   `json:"id"`
			DisplayName     string   `json:"display_name,omitempty"`
			Size            int      `json:"size"`
			Members         []string `json:"members"`
			Purity          float64  `json:"purity,omitempty"`
			CouplingRatio   float64  `json:"coupling_ratio,omitempty"`
			Betweenness     float64  `json:"betweenness,omitempty"`
			InboundEdges    int      `json:"inbound_edges"`
			OutboundEdges   int      `json:"outbound_edges"`
			BusFactor       int      `json:"bus_factor,omitempty"`
			TopNamespaces   []string `json:"top_namespaces,omitempty"`
			TopClasses      []string `json:"top_classes,omitempty"`
		}

		var communities []communityDetail
		for label, members := range cm.Communities {
			cd := communityDetail{
				ID:      label,
				Size:    len(members),
				Members: members,
			}
			if cm.DisplayNamePerComm != nil {
				cd.DisplayName = cm.DisplayNamePerComm[label]
			}
			if cm.PurityPerCommunity != nil {
				cd.Purity = cm.PurityPerCommunity[label]
			}
			if cm.CouplingRatioPerComm != nil {
				cd.CouplingRatio = cm.CouplingRatioPerComm[label]
			}
			if cm.BetweennessPerComm != nil {
				cd.Betweenness = cm.BetweennessPerComm[label]
			}
			if cm.InboundEdgesPerComm != nil {
				cd.InboundEdges = cm.InboundEdgesPerComm[label]
			}
			if cm.OutboundEdgesPerComm != nil {
				cd.OutboundEdges = cm.OutboundEdgesPerComm[label]
			}
			if cm.BusFactorPerCommunity != nil {
				cd.BusFactor = cm.BusFactorPerCommunity[label]
			}
			if cm.TopNamespacesPerComm != nil {
				cd.TopNamespaces = cm.TopNamespacesPerComm[label]
			}
			if cm.TopClassesPerComm != nil {
				cd.TopClasses = cm.TopClassesPerComm[label]
			}
			communities = append(communities, cd)
		}

		// Inter-community edges
		type interEdge struct {
			From  string `json:"from"`
			To    string `json:"to"`
			Edges int    `json:"edges"`
		}
		var interEdges []interEdge
		for _, e := range cm.EdgesBetweenCommunities {
			interEdges = append(interEdges, interEdge{
				From:  e.From,
				To:    e.To,
				Edges: e.Edges,
			})
		}

		result := map[string]any{
			"communities_count": cm.CommunitiesCount,
			"avg_size":          cm.AvgSize,
			"max_size":          cm.MaxSize,
			"graph_density":     cm.GraphDensity,
			"modularity":        cm.ModularityQ,
			"communities":       communities,
			"inter_edges":       interEdges,
			"boundary_nodes":    cm.BoundaryNodes,
		}

		// Suggestions related to communities
		var suggestions []map[string]string
		for _, s := range agg.Combined.Suggestions {
			suggestions = append(suggestions, map[string]string{
				"summary":  s.Summary,
				"location": s.Location,
				"why":      s.Why,
			})
		}
		if len(suggestions) > 0 {
			result["suggestions"] = suggestions
		}

		return safeToolResultJSON(result)
	}
}
