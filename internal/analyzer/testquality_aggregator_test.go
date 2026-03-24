package analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/stretchr/testify/assert"
)

func TestTestQualityAggregator_EmptyInput(t *testing.T) {
	agg := newAggregated()
	tqa := NewTestQualityAggregator()
	tqa.Calculate(&agg)

	assert.NotNil(t, agg.TestQuality)
	assert.Equal(t, 0, agg.TestQuality.NbTestFiles)
	assert.Equal(t, 0, agg.TestQuality.NbProdFiles)
}

func TestTestQualityAggregator_NilAggregate(t *testing.T) {
	tqa := NewTestQualityAggregator()
	tqa.Calculate(nil) // should not panic
}

func TestTestQualityAggregator_NoTestFiles(t *testing.T) {
	agg := newAggregated()

	ccn := int32(10)
	prodFile := &pb.File{
		Path:                "src/main.go",
		ShortPath:           "main.go",
		ProgrammingLanguage: "Go",
		IsTest:              false,
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Qualified: "Main", Short: "Main"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Complexity: &pb.Complexity{Cyclomatic: &ccn},
							Coupling:   &pb.Coupling{Efferent: 3, Afferent: 2},
						},
					},
				},
			},
		},
	}
	agg.ConcernedFiles = []*pb.File{prodFile}

	tqa := NewTestQualityAggregator()
	tqa.Calculate(&agg)

	assert.NotNil(t, agg.TestQuality)
	assert.Equal(t, 0, agg.TestQuality.NbTestFiles)
	assert.Equal(t, 1, agg.TestQuality.NbProdFiles)
	assert.Equal(t, 1, agg.TestQuality.NbProdClasses)
	assert.Equal(t, 0, agg.TestQuality.NbTestedClasses)
	assert.Equal(t, float64(0), agg.TestQuality.TraceabilityPct)
	// The class should appear as orphan
	assert.GreaterOrEqual(t, len(agg.TestQuality.OrphanClasses), 1)
	assert.Equal(t, "Main", agg.TestQuality.OrphanClasses[0].ClassName)
}

func TestTestQualityAggregator_BasicTraceability(t *testing.T) {
	agg := newAggregated()

	ccn := int32(5)
	prodFile := &pb.File{
		Path:                "src/service.go",
		ShortPath:           "service.go",
		ProgrammingLanguage: "Go",
		IsTest:              false,
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Qualified: "Service", Short: "Service"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Complexity: &pb.Complexity{Cyclomatic: &ccn},
							Coupling:   &pb.Coupling{Efferent: 1, Afferent: 1},
						},
					},
				},
			},
		},
	}

	testFile := &pb.File{
		Path:                "src/service_test.go",
		ShortPath:           "service_test.go",
		ProgrammingLanguage: "Go",
		IsTest:              true,
		Stmts: &pb.Stmts{
			StmtExternalDependencies: []*pb.StmtExternalDependency{
				{
					ClassName: "Service",
					Namespace: "src",
					From:      "ServiceTest",
				},
			},
		},
	}

	agg.ConcernedFiles = []*pb.File{prodFile, testFile}

	tqa := NewTestQualityAggregator()
	tqa.Calculate(&agg)

	assert.NotNil(t, agg.TestQuality)
	assert.Equal(t, 1, agg.TestQuality.NbTestFiles)
	assert.Equal(t, 1, agg.TestQuality.NbProdFiles)
	assert.Equal(t, 1, agg.TestQuality.NbTestedClasses)
	assert.Equal(t, float64(100), agg.TestQuality.TraceabilityPct)
	assert.Equal(t, 0, len(agg.TestQuality.OrphanClasses))

	// Test file should have fan-out of 1, high isolation
	assert.Equal(t, 1, len(agg.TestQuality.TestFiles))
	assert.Equal(t, 1, agg.TestQuality.TestFiles[0].SUTFanOut)
	assert.Equal(t, 90.0, agg.TestQuality.TestFiles[0].IsolationScore)
	assert.Equal(t, "Isolated", agg.TestQuality.TestFiles[0].IsolationLabel)
}

func TestTestQualityAggregator_GodTestDetection(t *testing.T) {
	agg := newAggregated()

	// Create 6 prod classes
	var prodClasses []*pb.StmtClass
	for i := 0; i < 6; i++ {
		name := "Class" + string(rune('A'+i))
		prodClasses = append(prodClasses, &pb.StmtClass{
			Name:  &pb.Name{Qualified: name, Short: name},
			Stmts: &pb.Stmts{Analyze: &pb.Analyze{}},
		})
	}

	prodFile := &pb.File{
		Path:                "src/classes.go",
		ShortPath:           "classes.go",
		ProgrammingLanguage: "Go",
		IsTest:              false,
		Stmts: &pb.Stmts{
			StmtClass: prodClasses,
		},
	}

	// Create test file that touches all 6 classes (god test, >= 5)
	var deps []*pb.StmtExternalDependency
	for i := 0; i < 6; i++ {
		name := "Class" + string(rune('A'+i))
		deps = append(deps, &pb.StmtExternalDependency{
			ClassName: name,
			Namespace: "src",
			From:      "GodTest",
		})
	}

	testFile := &pb.File{
		Path:                "src/god_test.go",
		ShortPath:           "god_test.go",
		ProgrammingLanguage: "Go",
		IsTest:              true,
		Stmts: &pb.Stmts{
			StmtExternalDependencies: deps,
		},
	}

	agg.ConcernedFiles = []*pb.File{prodFile, testFile}

	tqa := NewTestQualityAggregator()
	tqa.Calculate(&agg)

	assert.NotNil(t, agg.TestQuality)
	assert.Equal(t, 1, len(agg.TestQuality.GodTests))
	assert.Equal(t, 6, agg.TestQuality.GodTests[0].SUTFanOut)
	assert.Equal(t, "Coupled", agg.TestQuality.GodTests[0].IsolationLabel)
}

func TestTestQualityAggregator_IsolationScoreTiers(t *testing.T) {
	tests := []struct {
		fanOut         int
		expectedLabel  string
		minScore       float64
		maxScore       float64
	}{
		{0, "Isolated", 100, 100},
		{1, "Isolated", 85, 95},
		{3, "Semi-isolated", 55, 75},
		{10, "Coupled", 0, 5},
	}

	for _, tt := range tests {
		agg := newAggregated()

		// Create prod classes
		var prodClasses []*pb.StmtClass
		var deps []*pb.StmtExternalDependency
		for i := 0; i < tt.fanOut; i++ {
			name := "ProdClass" + string(rune('A'+i))
			prodClasses = append(prodClasses, &pb.StmtClass{
				Name:  &pb.Name{Qualified: name, Short: name},
				Stmts: &pb.Stmts{Analyze: &pb.Analyze{}},
			})
			deps = append(deps, &pb.StmtExternalDependency{
				ClassName: name,
				Namespace: "src",
				From:      "TestClass",
			})
		}

		var files []*pb.File
		if len(prodClasses) > 0 {
			files = append(files, &pb.File{
				Path: "src/prod.go", ShortPath: "prod.go",
				ProgrammingLanguage: "Go", IsTest: false,
				Stmts: &pb.Stmts{StmtClass: prodClasses},
			})
		}
		files = append(files, &pb.File{
			Path: "src/test.go", ShortPath: "test.go",
			ProgrammingLanguage: "Go", IsTest: true,
			Stmts: &pb.Stmts{StmtExternalDependencies: deps},
		})
		agg.ConcernedFiles = files

		tqa := NewTestQualityAggregator()
		tqa.Calculate(&agg)

		assert.NotNil(t, agg.TestQuality)
		assert.Equal(t, 1, len(agg.TestQuality.TestFiles))
		score := agg.TestQuality.TestFiles[0].IsolationScore
		assert.GreaterOrEqual(t, score, tt.minScore, "fanOut=%d score=%f < min=%f", tt.fanOut, score, tt.minScore)
		assert.LessOrEqual(t, score, tt.maxScore, "fanOut=%d score=%f > max=%f", tt.fanOut, score, tt.maxScore)
		assert.Equal(t, tt.expectedLabel, agg.TestQuality.TestFiles[0].IsolationLabel)
	}
}

func TestTestQualityAggregator_OrphanDetection(t *testing.T) {
	agg := newAggregated()

	ccn1 := int32(20)
	ccn2 := int32(5)

	prodFile := &pb.File{
		Path:                "src/service.go",
		ShortPath:           "service.go",
		ProgrammingLanguage: "Go",
		IsTest:              false,
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Qualified: "ImportantService", Short: "ImportantService"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Complexity: &pb.Complexity{Cyclomatic: &ccn1},
							Coupling:   &pb.Coupling{Efferent: 5, Afferent: 3},
						},
					},
				},
				{
					Name: &pb.Name{Qualified: "SimpleHelper", Short: "SimpleHelper"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Complexity: &pb.Complexity{Cyclomatic: &ccn2},
							Coupling:   &pb.Coupling{Efferent: 0, Afferent: 0},
						},
					},
				},
			},
		},
	}

	agg.ConcernedFiles = []*pb.File{prodFile}

	tqa := NewTestQualityAggregator()
	tqa.Calculate(&agg)

	assert.NotNil(t, agg.TestQuality)
	assert.Equal(t, 2, len(agg.TestQuality.OrphanClasses))
	// ImportantService should be first (higher weight)
	assert.Equal(t, "ImportantService", agg.TestQuality.OrphanClasses[0].ClassName)
	assert.Greater(t, agg.TestQuality.OrphanClasses[0].Weight, agg.TestQuality.OrphanClasses[1].Weight)
}

func TestBuildTestQualityJSON(t *testing.T) {
	metrics := &TestQualityMetrics{
		GlobalIsolationScore: 75.5,
		IsolationLabel:       "Semi-isolated",
		NbTestFiles:          5,
		NbProdFiles:          10,
		NbProdClasses:        8,
		NbTestedClasses:      6,
		TraceabilityPct:      75.0,
		IsolationHistogram:   [5]int{1, 2, 0, 1, 1},
	}

	json := BuildTestQualityJSON(metrics)
	assert.Contains(t, json, "\"globalIsolationScore\":75.5")
	assert.Contains(t, json, "\"isolationLabel\":\"Semi-isolated\"")
	assert.Contains(t, json, "\"nbTestFiles\":5")
	assert.Contains(t, json, "\"traceabilityPct\":75")
	assert.Contains(t, json, "\"isolationHistogram\":[1,2,0,1,1]")
}

func TestTestQualityAggregator_FunctionTraceability(t *testing.T) {
	// Test that standalone functions (not in classes) are tracked for traceability.
	// This is critical for functional codebases (TypeScript, JS, Python).
	agg := newAggregated()

	ccn := int32(3)
	prodFile := &pb.File{
		Path:                "src/utils.ts",
		ShortPath:           "utils.ts",
		ProgrammingLanguage: "TypeScript",
		IsTest:              false,
		Stmts: &pb.Stmts{
			StmtNamespace: []*pb.StmtNamespace{
				{
					Name: &pb.Name{Qualified: "utils", Short: "utils"},
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{
							{
								Name: &pb.Name{Qualified: "formatDate", Short: "formatDate"},
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Complexity: &pb.Complexity{Cyclomatic: &ccn},
									},
								},
							},
							{
								Name: &pb.Name{Qualified: "parseJSON", Short: "parseJSON"},
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Complexity: &pb.Complexity{Cyclomatic: &ccn},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	testFile := &pb.File{
		Path:                "src/utils.test.ts",
		ShortPath:           "utils.test.ts",
		ProgrammingLanguage: "TypeScript",
		IsTest:              true,
		Stmts: &pb.Stmts{
			StmtExternalDependencies: []*pb.StmtExternalDependency{
				{
					ClassName: "formatDate",
					Namespace: "./utils",
					From:      "utils.test",
				},
			},
		},
	}

	agg.ConcernedFiles = []*pb.File{prodFile, testFile}

	tqa := NewTestQualityAggregator()
	tqa.Calculate(&agg)

	assert.NotNil(t, agg.TestQuality)
	assert.Equal(t, 1, agg.TestQuality.NbTestFiles)
	assert.Equal(t, 1, agg.TestQuality.NbProdFiles)
	// 2 prod structures: formatDate + parseJSON
	assert.Equal(t, 2, agg.TestQuality.NbProdClasses)
	// 1 tested (formatDate)
	assert.Equal(t, 1, agg.TestQuality.NbTestedClasses)
	assert.Equal(t, float64(50), agg.TestQuality.TraceabilityPct)
	// parseJSON should be orphan
	assert.Equal(t, 1, len(agg.TestQuality.OrphanClasses))
	assert.Equal(t, "parseJSON", agg.TestQuality.OrphanClasses[0].ClassName)
}

func TestTestQualityAggregator_MixedClassesAndFunctions(t *testing.T) {
	// Test that both classes and standalone functions contribute to traceability
	agg := newAggregated()

	ccn := int32(5)
	prodFile := &pb.File{
		Path:                "src/app.ts",
		ShortPath:           "app.ts",
		ProgrammingLanguage: "TypeScript",
		IsTest:              false,
		Stmts: &pb.Stmts{
			StmtClass: []*pb.StmtClass{
				{
					Name: &pb.Name{Qualified: "AppService", Short: "AppService"},
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Complexity: &pb.Complexity{Cyclomatic: &ccn},
						},
					},
				},
			},
			StmtNamespace: []*pb.StmtNamespace{
				{
					Name: &pb.Name{Qualified: "app", Short: "app"},
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{
							{
								Name: &pb.Name{Qualified: "initApp", Short: "initApp"},
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Complexity: &pb.Complexity{Cyclomatic: &ccn},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	testFile := &pb.File{
		Path:                "src/app.test.ts",
		ShortPath:           "app.test.ts",
		ProgrammingLanguage: "TypeScript",
		IsTest:              true,
		Stmts: &pb.Stmts{
			StmtExternalDependencies: []*pb.StmtExternalDependency{
				{ClassName: "AppService", Namespace: "./app", From: "app.test"},
				{ClassName: "initApp", Namespace: "./app", From: "app.test"},
			},
		},
	}

	agg.ConcernedFiles = []*pb.File{prodFile, testFile}

	tqa := NewTestQualityAggregator()
	tqa.Calculate(&agg)

	assert.NotNil(t, agg.TestQuality)
	// 2 prod structures: AppService (class) + initApp (function)
	assert.Equal(t, 2, agg.TestQuality.NbProdClasses)
	// Both tested
	assert.Equal(t, 2, agg.TestQuality.NbTestedClasses)
	assert.Equal(t, float64(100), agg.TestQuality.TraceabilityPct)
	assert.Equal(t, 0, len(agg.TestQuality.OrphanClasses))
}

func TestBuildTestQualityJSON_Nil(t *testing.T) {
	json := BuildTestQualityJSON(nil)
	assert.Equal(t, "{}", json)
}
