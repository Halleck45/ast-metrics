package mcp

import (
	"testing"
	"time"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/stretchr/testify/assert"
)

func TestCacheGetEmpty(t *testing.T) {
	cache := NewAnalysisCache(60 * time.Second)
	files, agg, ok := cache.Get()
	assert.False(t, ok)
	assert.Nil(t, files)
	assert.Nil(t, agg)
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewAnalysisCache(60 * time.Second)

	files := []*pb.File{{Path: "test.go"}}
	agg := analyzer.ProjectAggregated{
		Combined: analyzer.Aggregated{NbFiles: 1},
	}

	cache.Set(files, agg)

	gotFiles, gotAgg, ok := cache.Get()
	assert.True(t, ok)
	assert.Equal(t, 1, len(gotFiles))
	assert.Equal(t, "test.go", gotFiles[0].Path)
	assert.Equal(t, 1, gotAgg.Combined.NbFiles)
}

func TestCacheExpiration(t *testing.T) {
	cache := NewAnalysisCache(1 * time.Millisecond)

	files := []*pb.File{{Path: "test.go"}}
	agg := analyzer.ProjectAggregated{}

	cache.Set(files, agg)

	time.Sleep(5 * time.Millisecond)

	_, _, ok := cache.Get()
	assert.False(t, ok)
}

func TestCacheInvalidate(t *testing.T) {
	cache := NewAnalysisCache(60 * time.Second)

	cache.Set([]*pb.File{{Path: "test.go"}}, analyzer.ProjectAggregated{})

	cache.Invalidate()

	_, _, ok := cache.Get()
	assert.False(t, ok)
}
