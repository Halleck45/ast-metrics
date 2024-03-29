package Analyzer

import (
	"reflect"
	"testing"

	"github.com/halleck45/ast-metrics/src/Analyzer"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
)

func TestBusFactorCalculate(t *testing.T) {
	tests := []struct {
		name     string
		files    []*pb.File
		expected Analyzer.Aggregated
	}{
		{
			name:  "Test with no commits",
			files: []*pb.File{},
			expected: Analyzer.Aggregated{
				TopCommitters: []Analyzer.TopCommitter{},
				BusFactor:     0,
			},
		},
		{
			name: "Test with one committer",
			files: []*pb.File{
				{
					Commits: &pb.Commits{
						Commits: []*pb.Commit{
							{Author: "author1"},
						},
					},
				},
			},
			expected: Analyzer.Aggregated{
				TopCommitters: []Analyzer.TopCommitter{
					{Name: "author1", Count: 1},
				},
				BusFactor: 1,
			},
		},
		{
			name: "Test with multiple committers",
			files: []*pb.File{
				{
					Commits: &pb.Commits{
						Commits: []*pb.Commit{
							{Author: "author1"},
							{Author: "author2"},
							{Author: "author1"},
						},
					},
				},
			},
			expected: Analyzer.Aggregated{
				TopCommitters: []Analyzer.TopCommitter{
					{Name: "author1", Count: 2},
					{Name: "author2", Count: 1},
				},
				BusFactor: 1,
			},
		},
		{
			name: "Test with excluded committers",
			files: []*pb.File{
				{
					Commits: &pb.Commits{
						Commits: []*pb.Commit{
							{Author: "author1"},
							{Author: "noreply@github.com"},
							{Author: ""},
						},
					},
				},
			},
			expected: Analyzer.Aggregated{
				TopCommitters: []Analyzer.TopCommitter{
					{Name: "author1", Count: 1},
				},
				BusFactor: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			busFactor := &BusFactor{}
			aggregate := &Analyzer.Aggregated{
				ConcernedFiles: tt.files,
			}
			busFactor.Calculate(aggregate)

			if !reflect.DeepEqual(aggregate.TopCommitters, tt.expected.TopCommitters) {
				t.Errorf("TopCommitters = %v, want %v", aggregate.TopCommitters, tt.expected.TopCommitters)
			}

			if aggregate.BusFactor != tt.expected.BusFactor {
				t.Errorf("BusFactor = %v, want %v", aggregate.BusFactor, tt.expected.BusFactor)
			}
		})
	}
}
