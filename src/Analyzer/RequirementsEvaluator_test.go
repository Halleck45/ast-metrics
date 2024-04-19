package Analyzer

import (
	"testing"

	"github.com/halleck45/ast-metrics/src/Configuration"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/stretchr/testify/assert"
)

func TestEvaluationResult(t *testing.T) {

	ccn5 := int32(5)
	ccn10 := int32(10)
	files := []*pb.File{
		{
			Path: "test1.go",
			Stmts: &pb.Stmts{
				Analyze: &pb.Analyze{
					Complexity: &pb.Complexity{
						Cyclomatic: &ccn10,
					},
				},
			},
		},
		{
			Path: "test2.go",
			Stmts: &pb.Stmts{
				Analyze: &pb.Analyze{
					Complexity: &pb.Complexity{
						Cyclomatic: &ccn5,
					},
				},
			},
		},
	}

	configInYaml := `
requirements:
  rules:
    cyclomatic_complexity:
      max: 5
`

	loader := Configuration.NewConfigurationLoader()
	configuration, err := loader.Import(configInYaml)
	assert.Nil(t, err)

	evaluator := NewRequirementsEvaluator(*configuration.Requirements)
	evaluation := evaluator.Evaluate(files, ProjectAggregated{})

	assert.Equal(t, files, evaluation.Files)
	assert.Equal(t, false, evaluation.Succeeded)

	assert.Equal(t, 1, len(evaluation.Errors))
	assert.Equal(t, "Cyclomatic complexity too high in file test1.go: got 10 (max: 5)", evaluation.Errors[0])
}

func TestEvaluationResultSuccess(t *testing.T) {

	ccn5 := int32(5)
	ccn10 := int32(10)
	files := []*pb.File{
		{
			Path: "test1.go",
			Stmts: &pb.Stmts{
				Analyze: &pb.Analyze{
					Complexity: &pb.Complexity{
						Cyclomatic: &ccn10,
					},
				},
			},
		},
		{
			Path: "test2.go",
			Stmts: &pb.Stmts{
				Analyze: &pb.Analyze{
					Complexity: &pb.Complexity{
						Cyclomatic: &ccn5,
					},
				},
			},
		},
	}

	configInYaml := `
requirements:
  rules:
    cyclomatic_complexity:
      max: 15
`

	loader := Configuration.NewConfigurationLoader()
	configuration, err := loader.Import(configInYaml)
	assert.Nil(t, err)

	evaluator := NewRequirementsEvaluator(*configuration.Requirements)
	evaluation := evaluator.Evaluate(files, ProjectAggregated{})

	assert.Equal(t, files, evaluation.Files)
	assert.Equal(t, true, evaluation.Succeeded)

	assert.Equal(t, 0, len(evaluation.Errors))
}
