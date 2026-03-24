package mcp

import (
	"sync"
	"time"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	Activity "github.com/halleck45/ast-metrics/internal/analyzer/activity"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
)

// AnalysisService wraps the ast-metrics analysis pipeline for reuse by MCP tools.
type AnalysisService struct {
	config  *configuration.Configuration
	runners []engine.Engine
	cache   *AnalysisCache
	mu      sync.Mutex
}

// NewAnalysisService creates a new service with the given configuration and runners.
func NewAnalysisService(config *configuration.Configuration, runners []engine.Engine) *AnalysisService {
	return &AnalysisService{
		config:  config,
		runners: runners,
		cache:   NewAnalysisCache(60 * time.Second),
	}
}

// Analyze runs the full analysis pipeline, using cached results when available.
// Set forceRefresh to bypass the cache.
func (s *AnalysisService) Analyze(forceRefresh bool) (*analyzer.ProjectAggregated, []*pb.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !forceRefresh {
		if files, agg, ok := s.cache.Get(); ok {
			return agg, files, nil
		}
	}

	// 1. Parse source files
	parsedFiles, err := engine.ParseFiles(s.config, s.runners)
	if err != nil {
		return nil, nil, err
	}

	// 2. Run metric analysis
	allResults := analyzer.AnalyzeFiles(parsedFiles, nil)

	// 3. Git analysis
	gitAnalyzer := analyzer.NewGitAnalyzer()
	gitSummaries := gitAnalyzer.Start(allResults)

	// 4. Aggregate results
	aggregator := analyzer.NewAggregator(allResults, gitSummaries)
	aggregator.WithAggregateAnalyzer(Activity.NewBusFactor())
	projectAggregated := aggregator.Aggregates()

	// 5. Risk analysis
	riskAnalyzer := analyzer.NewRiskAnalyzer()
	riskAnalyzer.Analyze(projectAggregated)

	// Cache and return
	s.cache.Set(allResults, projectAggregated)

	return &projectAggregated, allResults, nil
}
