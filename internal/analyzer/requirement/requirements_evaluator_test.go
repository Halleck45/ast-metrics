package requirement

import (
	"fmt"
	"testing"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
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
    complexity:
      max_cyclomatic: 5`

	loader := configuration.NewConfigurationLoader()
	config, err := loader.Import(configInYaml)
	assert.Nil(t, err)
	fmt.Println(config)

	evaluator := NewRequirementsEvaluator(*config.Requirements)
	evaluation := evaluator.Evaluate(files, ProjectAggregated{})

	assert.Equal(t, files, evaluation.Files)
	assert.Equal(t, false, evaluation.Succeeded)

	assert.Equal(t, 1, len(evaluation.Errors))
	assert.Equal(t, "Cyclomatic complexity too high: got 10 (max: 5)", evaluation.Errors[0].Message)
	assert.Equal(t, "cyclomatic_complexity", evaluation.Errors[0].Rule)
	assert.Equal(t, "test1.go", evaluation.Errors[0].File)
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
    max_cyclomatic: 15
`

	loader := configuration.NewConfigurationLoader()
	configuration, err := loader.Import(configInYaml)
	assert.Nil(t, err)

	evaluator := NewRequirementsEvaluator(*configuration.Requirements)
	evaluation := evaluator.Evaluate(files, ProjectAggregated{})

	assert.Equal(t, files, evaluation.Files)
	assert.Equal(t, true, evaluation.Succeeded)

	assert.Equal(t, 0, len(evaluation.Errors))
}
