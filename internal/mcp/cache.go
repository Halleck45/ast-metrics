package mcp

import (
	"sync"
	"time"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
)

// AnalysisCache stores analysis results with TTL-based invalidation.
type AnalysisCache struct {
	mu          sync.Mutex
	result      *cachedResult
	ttl         time.Duration
}

type cachedResult struct {
	files       []*pb.File
	aggregated  analyzer.ProjectAggregated
	createdAt   time.Time
}

// NewAnalysisCache creates a cache with the given TTL.
func NewAnalysisCache(ttl time.Duration) *AnalysisCache {
	return &AnalysisCache{
		ttl: ttl,
	}
}

// Get returns cached results if they exist and haven't expired.
func (c *AnalysisCache) Get() ([]*pb.File, *analyzer.ProjectAggregated, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.result == nil {
		return nil, nil, false
	}

	if time.Since(c.result.createdAt) > c.ttl {
		c.result = nil
		return nil, nil, false
	}

	return c.result.files, &c.result.aggregated, true
}

// Set stores analysis results in the cache.
func (c *AnalysisCache) Set(files []*pb.File, aggregated analyzer.ProjectAggregated) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.result = &cachedResult{
		files:      files,
		aggregated: aggregated,
		createdAt:  time.Now(),
	}
}

// Invalidate clears the cache.
func (c *AnalysisCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.result = nil
}
