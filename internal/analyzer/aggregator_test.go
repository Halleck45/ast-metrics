package analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestConsolidate(t *testing.T) {

	aggregator := Aggregator{}
	aggregated := Aggregated{
		MethodsPerClass:                     AggregateResult{Sum: 10, Counter: 5},
		NbClasses:                           5,
		NbClassesWithCode:                   5,
		CyclomaticComplexityPerClass:        AggregateResult{Sum: 20, Counter: 5},
		HalsteadDifficulty:                  AggregateResult{Sum: 30, Counter: 5},
		HalsteadEffort:                      AggregateResult{Sum: 40, Counter: 5},
		HalsteadVolume:                      AggregateResult{Sum: 50, Counter: 5},
		HalsteadTime:                        AggregateResult{Sum: 60, Counter: 5},
		LocPerMethod:                        AggregateResult{Sum: 70, Counter: 10},
		ClocPerMethod:                       AggregateResult{Sum: 80, Counter: 10},
		LlocPerMethod:                       AggregateResult{Sum: 90, Counter: 10},
		MaintainabilityIndex:                AggregateResult{Sum: 100, Counter: 5},
		MaintainabilityIndexWithoutComments: AggregateResult{Sum: 110, Counter: 5},
	}

	aggregated = aggregator.reduceMetrics(aggregated)

	assert.Equal(t, float64(2), aggregated.MethodsPerClass.Avg, "Should have 2 methods per class")
	assert.Equal(t, float64(10), aggregated.MethodsPerClass.Sum, "Should have 10 methods per class sum")
	assert.Equal(t, float64(4), aggregated.CyclomaticComplexityPerClass.Avg, "Should have 4 cyclomatic complexity per class")
	assert.Equal(t, float64(6), aggregated.HalsteadDifficulty.Avg, "Should have 6 halstead difficulty")
	assert.Equal(t, float64(8), aggregated.HalsteadEffort.Avg, "Should have 8 halstead effort")
	assert.Equal(t, float64(10), aggregated.HalsteadVolume.Avg, "Should have 10 halstead volume")
	assert.Equal(t, float64(12), aggregated.HalsteadTime.Avg, "Should have 12 halstead time")
	assert.Equal(t, float64(7), aggregated.LocPerMethod.Avg, "Should have 7 loc per method")
	assert.Equal(t, float64(8), aggregated.ClocPerMethod.Avg, "Should have 8 cloc per method")
	assert.Equal(t, float64(9), aggregated.LlocPerMethod.Avg, "Should have 9 lloc per method")
	assert.Equal(t, float64(20), aggregated.MaintainabilityIndex.Avg, "Should have 20 maintainability index")
	assert.Equal(t, float64(22), aggregated.MaintainabilityIndexWithoutComments.Avg, "Should have 22 maintainability index without comments")

}

func TestCalculate(t *testing.T) {

	t.Run("TestCalculate", func(t *testing.T) {
		aggregator := Aggregator{}
		stmts := pb.Stmts{
			StmtFunction: []*pb.StmtFunction{
				{
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Complexity: &pb.Complexity{
								Cyclomatic: proto.Int32(10),
							},
						},
					},
				},
				{
					Stmts: &pb.Stmts{
						Analyze: &pb.Analyze{
							Complexity: &pb.Complexity{
								Cyclomatic: proto.Int32(20),
							},
						},
					},
				},
			},
			StmtClass: []*pb.StmtClass{
				{}, {}, {},
			},
			Analyze: &pb.Analyze{
				Volume: &pb.Volume{
					Loc:  proto.Int32(100),
					Cloc: proto.Int32(200),
					Lloc: proto.Int32(300),
				},
			},
		}
		file := pb.File{
			Stmts: &stmts,
			Path:  "test.foo",
		}
		aggregated := Aggregated{}
		aggregated = aggregator.mapSums(&file, aggregated)
		aggregated.ConcernedFiles = []*pb.File{
			&file,
		}
		aggregated = aggregator.reduceMetrics(aggregated)

		assert.Equal(t, 2, aggregated.NbMethods, "Should have 2 methods")
		assert.Equal(t, 3, aggregated.NbClasses, "Should have 3 classes")
		assert.Equal(t, float64(15), aggregated.CyclomaticComplexityPerMethod.Avg, "Should have 15 average cyclomatic complexity per method")
		assert.Equal(t, float64(100), aggregated.Loc.Avg, "Should have 100 loc")
		assert.Equal(t, float64(200), aggregated.Cloc.Avg, "Should have 200 cloc")
		assert.Equal(t, float64(300), aggregated.Lloc.Avg, "Should have 300 lloc")
	})
}

func TestAggregates(t *testing.T) {
	t.Run("TestAggregates", func(t *testing.T) {
		// Create a new Aggregator with some dummy data
		aggregator := Aggregator{
			files: []*pb.File{
				// file 1
				{
					ProgrammingLanguage: "Go",
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{
							{
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Complexity: &pb.Complexity{
											Cyclomatic: proto.Int32(10),
										},
									},
								},
							},
						},
						StmtClass: []*pb.StmtClass{
							// class
							{
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Maintainability: &pb.Maintainability{
											MaintainabilityIndex:                proto.Float64(120),
											MaintainabilityIndexWithoutComments: proto.Float64(48),
											CommentWeight:                       proto.Float64(40),
										},
									},
								},
							},
							// class
							{
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Maintainability: &pb.Maintainability{
											MaintainabilityIndex:                proto.Float64(85),
											MaintainabilityIndexWithoutComments: proto.Float64(48),
											CommentWeight:                       proto.Float64(40),
										},
									},
								},
							},
							// class
							{
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Maintainability: &pb.Maintainability{
											MaintainabilityIndex:                proto.Float64(65),
											MaintainabilityIndexWithoutComments: proto.Float64(48),
											CommentWeight:                       proto.Float64(40),
										},
									},
								},
							},
							// class
							{
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Maintainability: &pb.Maintainability{
											MaintainabilityIndex:                proto.Float64(100),
											MaintainabilityIndexWithoutComments: proto.Float64(48),
											CommentWeight:                       proto.Float64(40),
										},
									},
								},
							},
						},
						StmtNamespace: []*pb.StmtNamespace{
							{
								Stmts: &pb.Stmts{
									StmtFunction: []*pb.StmtFunction{
										{
											Stmts: &pb.Stmts{
												Analyze: &pb.Analyze{
													Complexity: &pb.Complexity{
														Cyclomatic: proto.Int32(20),
													},
												},
											},
										},
									},
									StmtClass: []*pb.StmtClass{
										// class
										{
											Stmts: &pb.Stmts{
												Analyze: &pb.Analyze{
													Maintainability: &pb.Maintainability{
														MaintainabilityIndex:                proto.Float64(70),
														MaintainabilityIndexWithoutComments: proto.Float64(48),
														CommentWeight:                       proto.Float64(40),
													},
												},
											},
										},
										// class
										{
											Stmts: &pb.Stmts{
												Analyze: &pb.Analyze{
													Maintainability: &pb.Maintainability{
														MaintainabilityIndex:                proto.Float64(100),
														MaintainabilityIndexWithoutComments: proto.Float64(48),
														CommentWeight:                       proto.Float64(40),
													},
												},
											},
										},
									},
								},
							},
						},
						Analyze: &pb.Analyze{
							Volume: &pb.Volume{
								Loc:  proto.Int32(100),
								Cloc: proto.Int32(200),
								Lloc: proto.Int32(50),
							},
						},
					},
				},
				// file 2
				{
					ProgrammingLanguage: "Go",
					Stmts: &pb.Stmts{
						StmtFunction: []*pb.StmtFunction{
							{
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Complexity: &pb.Complexity{
											Cyclomatic: proto.Int32(60),
										},
									},
								},
							},
						},
						StmtClass: []*pb.StmtClass{
							// class
							{
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Maintainability: &pb.Maintainability{
											MaintainabilityIndex:                proto.Float64(75),
											MaintainabilityIndexWithoutComments: proto.Float64(48),
											CommentWeight:                       proto.Float64(40),
										},
									},
								},
							},
							// class
							{
								Stmts: &pb.Stmts{
									Analyze: &pb.Analyze{
										Maintainability: &pb.Maintainability{
											MaintainabilityIndex:                proto.Float64(120),
											MaintainabilityIndexWithoutComments: proto.Float64(48),
											CommentWeight:                       proto.Float64(40),
										},
									},
								},
							},
						},
						StmtNamespace: []*pb.StmtNamespace{
							{
								Stmts: &pb.Stmts{
									StmtFunction: []*pb.StmtFunction{
										{
											Stmts: &pb.Stmts{
												Analyze: &pb.Analyze{
													Complexity: &pb.Complexity{
														Cyclomatic: proto.Int32(30),
													},
												},
											},
										},
									},
									StmtClass: []*pb.StmtClass{
										// class
										{
											Stmts: &pb.Stmts{
												Analyze: &pb.Analyze{
													Maintainability: &pb.Maintainability{
														MaintainabilityIndex:                proto.Float64(90),
														MaintainabilityIndexWithoutComments: proto.Float64(48),
														CommentWeight:                       proto.Float64(40),
													},
												},
											},
										},
									},
								},
							},
						},
						Analyze: &pb.Analyze{
							Volume: &pb.Volume{
								Loc:  proto.Int32(200),
								Cloc: proto.Int32(300),
								Lloc: proto.Int32(150),
							},
						},
					},
				},
				// file 3
				{
					ProgrammingLanguage: "Php",
					Stmts: &pb.Stmts{
						StmtNamespace: []*pb.StmtNamespace{
							{
								Stmts: &pb.Stmts{
									StmtFunction: []*pb.StmtFunction{
										{
											Stmts: &pb.Stmts{
												Analyze: &pb.Analyze{
													Complexity: &pb.Complexity{
														Cyclomatic: proto.Int32(30),
													},
												},
											},
										},
									},
									StmtClass: []*pb.StmtClass{
										// class
										{
											Stmts: &pb.Stmts{
												Analyze: &pb.Analyze{
													Maintainability: &pb.Maintainability{
														MaintainabilityIndex:                proto.Float64(120),
														MaintainabilityIndexWithoutComments: proto.Float64(48),
														CommentWeight:                       proto.Float64(40),
													},
												},
											},
										},
									},
								},
							},
						},
						Analyze: &pb.Analyze{
							Volume: &pb.Volume{
								Loc:  proto.Int32(600),
								Cloc: proto.Int32(100),
								Lloc: proto.Int32(400),
							},
						},
					},
				},
			},
		}

		// Call the Aggregates method
		projectAggregated := aggregator.Aggregates()
		result := projectAggregated.Combined

		// Check that the returned ProjectAggregated struct has the expected values
		assert.Equal(t, 3, result.NbFiles, "Should have 3 files")

		// Checks on Combined aggregate
		assert.Equal(t, 10, projectAggregated.ByClass.NbClasses, "Should have 10 classes")

		assert.Equal(t, 5, result.NbMethods, "Should have 5 methods")

		assert.Equal(t, float64(30), result.CyclomaticComplexityPerMethod.Avg, "Should have 30 average cyclomatic complexity per method")

		assert.Equal(t, 94, int(result.MaintainabilityIndex.Avg), "Should have 94 average maintainability index")

		// Check on Go aggregate
		assert.Equal(t, 9, projectAggregated.ByProgrammingLanguage["Go"].NbClasses, "Should have 9 classes")

		assert.Equal(t, 4, projectAggregated.ByProgrammingLanguage["Go"].NbMethods, "Should have 4 methods in Go")

		assert.Equal(t, 2, projectAggregated.ByProgrammingLanguage["Go"].NbFiles, "Should have 2 Go files")

		assert.Equal(t, 91, int(projectAggregated.ByProgrammingLanguage["Go"].MaintainabilityIndex.Avg), "Should have 91 average maintainability index for Go files")

		// Check on Php aggregate
		assert.Equal(t, 1, projectAggregated.ByProgrammingLanguage["Php"].NbClasses, "Should have 1 class")

		assert.Equal(t, 1, projectAggregated.ByProgrammingLanguage["Php"].NbMethods, "Should have 1 methods in PHP")

		assert.Equal(t, 1, projectAggregated.ByProgrammingLanguage["Php"].NbFiles, "Should have 1 PHP files")

		assert.Equal(t, 120, int(projectAggregated.ByProgrammingLanguage["Php"].MaintainabilityIndex.Avg), "Should have 120 average maintainability index for PHP files")
	})
}

func TestCalculateMaintainabilityIndex(t *testing.T) {
	t.Run("TestCalculateMaintainabilityIndex", func(t *testing.T) {
		aggregator := Aggregator{}
		file := pb.File{
			Stmts: &pb.Stmts{
				StmtFunction: []*pb.StmtFunction{
					{
						Stmts: &pb.Stmts{
							Analyze: &pb.Analyze{
								Maintainability: &pb.Maintainability{
									MaintainabilityIndex:                proto.Float64(15),
									MaintainabilityIndexWithoutComments: proto.Float64(20),
									CommentWeight:                       proto.Float64(25),
								},
							},
						},
					},
					{
						Stmts: &pb.Stmts{
							Analyze: &pb.Analyze{
								Maintainability: &pb.Maintainability{
									MaintainabilityIndex:                proto.Float64(30),
									MaintainabilityIndexWithoutComments: proto.Float64(35),
									CommentWeight:                       proto.Float64(40),
								},
							},
						},
					},
				},
			},
		}
		aggregated := Aggregated{}

		aggregated = aggregator.mapSums(&file, aggregated)
		aggregated = aggregator.reduceMetrics(aggregated)

		assert.Equal(t, float64(22.5), aggregated.MaintainabilityIndex.Avg, "Should have 22.5 average maintainability index")
		assert.Equal(t, float64(27.5), aggregated.MaintainabilityIndexWithoutComments.Avg, "Should have 27.5 average maintainability index without comments")
		assert.Equal(t, float64(22.5), aggregated.MaintainabilityPerMethod.Avg, "Should have 22.5 average maintainability index per method")
	})
}

func TestFIlesWithErrorAreDetected(t *testing.T) {
	t.Run("TestFilesWithErrorAreDetected", func(t *testing.T) {
		aggregator := Aggregator{}
		files := []*pb.File{
			&pb.File{
				Stmts: &pb.Stmts{},
			},
			&pb.File{
				Errors: []string{"Error1", "Error2"},
			},
		}
		aggregator.files = files
		aggregated := aggregator.Aggregates()

		assert.Equal(t, 2, aggregated.ByFile.NbFiles)
		assert.Equal(t, 1, len(aggregated.ErroredFiles))
	})
}
