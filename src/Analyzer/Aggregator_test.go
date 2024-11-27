package Analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestConsolidate(t *testing.T) {

	aggregator := Aggregator{}
	aggregated := Aggregated{
		NbMethods:                           10,
		NbClasses:                           5,
		NbClassesWithCode:                   5,
		AverageCyclomaticComplexityPerClass: 20,
		AverageHalsteadDifficulty:           30,
		AverageHalsteadEffort:               40,
		AverageHalsteadVolume:               50,
		AverageHalsteadTime:                 60,
		AverageLocPerMethod:                 70,
		AverageClocPerMethod:                80,
		AverageLlocPerMethod:                90,
		AverageMI:                           100,
		AverageMIwoc:                        110,
		AverageMIcw:                         120,
	}

	aggregator.consolidate(&aggregated)

	if aggregated.AverageMethodsPerClass != 2 {
		t.Errorf("Expected 2, got %f", aggregated.AverageMethodsPerClass)
	}

	if aggregated.AverageCyclomaticComplexityPerClass != 4 {
		t.Errorf("Expected 4, got %f", aggregated.AverageCyclomaticComplexityPerClass)
	}

	if aggregated.AverageHalsteadDifficulty != 6 {
		t.Errorf("Expected 6, got %f", aggregated.AverageHalsteadDifficulty)
	}

	if aggregated.AverageHalsteadEffort != 8 {
		t.Errorf("Expected 8, got %f", aggregated.AverageHalsteadEffort)
	}

	if aggregated.AverageHalsteadVolume != 10 {
		t.Errorf("Expected 10, got %f", aggregated.AverageHalsteadVolume)
	}

	if aggregated.AverageHalsteadTime != 12 {
		t.Errorf("Expected 12, got %f", aggregated.AverageHalsteadTime)
	}

	if aggregated.AverageLocPerMethod != 7 {
		t.Errorf("Expected 7, got %f", aggregated.AverageLocPerMethod)
	}

	if aggregated.AverageClocPerMethod != 8 {
		t.Errorf("Expected 8, got %f", aggregated.AverageClocPerMethod)
	}

	if aggregated.AverageLlocPerMethod != 9 {
		t.Errorf("Expected 9, got %f", aggregated.AverageLlocPerMethod)
	}

	if aggregated.AverageMI != 20 {
		t.Errorf("Expected 20, got %f", aggregated.AverageMI)
	}

	if aggregated.AverageMIwoc != 22 {
		t.Errorf("Expected 22, got %f", aggregated.AverageMIwoc)
	}

	if aggregated.AverageMIcw != 24 {
		t.Errorf("Expected 24, got %f", aggregated.AverageMIcw)
	}
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
		aggregator.calculateSums(&file, &aggregated)
		aggregated.ConcernedFiles = []*pb.File{
			&file,
		}
		aggregator.consolidate(&aggregated)

		if aggregated.NbMethods != 2 {
			t.Errorf("Expected 2, got %d", aggregated.NbMethods)
		}

		if aggregated.NbClasses != 3 {
			t.Errorf("Expected 3 classes, got %d", aggregated.NbClasses)
		}

		if aggregated.AverageCyclomaticComplexityPerMethod != 15 {
			t.Errorf("Expected AverageCyclomaticComplexityPerMethod, got %f", aggregated.AverageCyclomaticComplexityPerMethod)
		}

		if aggregated.Loc != 100 {
			t.Errorf("Expected 100, got %d", aggregated.Loc)
		}

		if aggregated.Cloc != 200 {
			t.Errorf("Expected 200, got %d", aggregated.Cloc)
		}

		if aggregated.Lloc != 300 {
			t.Errorf("Expected 300, got %d", aggregated.Lloc)
		}
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

		// Check that the returned ProjectAggregated struct has the expected values
		if projectAggregated.ByFile.NbFiles != 3 {
			t.Errorf("Expected 3 files, got %d", projectAggregated.ByFile.NbFiles)
		}

		// Checks on Combined aggregate
		if projectAggregated.ByClass.NbClasses != 10 {
			t.Errorf("Expected 10 classes, got %d", projectAggregated.ByClass.NbClasses)
		}

		if projectAggregated.Combined.NbClasses != 10 {
			t.Errorf("Expected 10 classes, got %d", projectAggregated.ByClass.NbClasses)
		}

		if projectAggregated.Combined.NbMethods != 5 {
			t.Errorf("Expected 5 methods, got %d", projectAggregated.Combined.NbMethods)
		}

		if projectAggregated.Combined.AverageCyclomaticComplexityPerMethod != 30 {
			t.Errorf("Expected AverageCyclomaticComplexityPerMethod 30, got %f", projectAggregated.Combined.AverageCyclomaticComplexityPerMethod)
		}

		if int(projectAggregated.Combined.AverageMI) != 94 {
			t.Errorf("Expected MI of 94 for all files, got %v", int(projectAggregated.Combined.AverageMI))
		}

		// Check on Go aggregate
		if projectAggregated.ByProgrammingLanguage["Go"].NbClasses != 9 {
			t.Errorf("Expected 9 classes, got %d", projectAggregated.ByProgrammingLanguage["Go"].NbClasses)
		}

		if projectAggregated.ByProgrammingLanguage["Go"].NbMethods != 4 {
			t.Errorf("Expected 4 methods in Go, got %d", projectAggregated.ByProgrammingLanguage["Go"].NbMethods)
		}

		if projectAggregated.ByProgrammingLanguage["Go"].NbFiles != 2 {
			t.Errorf("Expected 2 Go files, got %d", projectAggregated.ByProgrammingLanguage["Go"].NbFiles)
		}

		if int(projectAggregated.ByProgrammingLanguage["Go"].AverageMI) != 91 {
			t.Errorf("Expected MI of 91 for Go files, got %v", int(projectAggregated.ByProgrammingLanguage["Go"].AverageMI))
		}

		// Check on Php aggregate
		if projectAggregated.ByProgrammingLanguage["Php"].NbClasses != 1 {
			t.Errorf("Expected 1 class, got %d", projectAggregated.ByProgrammingLanguage["Php"].NbClasses)
		}

		if projectAggregated.ByProgrammingLanguage["Php"].NbMethods != 1 {
			t.Errorf("Expected 1 methods in PHP, got %d", projectAggregated.ByProgrammingLanguage["Php"].NbMethods)
		}

		if projectAggregated.ByProgrammingLanguage["Php"].NbFiles != 1 {
			t.Errorf("Expected 1 PHP files, got %d", projectAggregated.ByProgrammingLanguage["Go"].NbFiles)
		}

		if projectAggregated.ByProgrammingLanguage["Php"].AverageMI != 120 {
			t.Errorf("Expected MI of 120 for PHP files, got %f", projectAggregated.ByProgrammingLanguage["Php"].AverageMI)
		}

		if int(projectAggregated.ByProgrammingLanguage["Php"].AverageMI) != 120 {
			t.Errorf("Expected MI of 120 for PHP files, got %v", int(projectAggregated.ByProgrammingLanguage["Go"].AverageMI))
		}
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

		aggregator.calculateSums(&file, &aggregated)
		aggregator.consolidate(&aggregated)

		if aggregated.AverageMI != 22.5 {
			t.Errorf("Expected 22.5, got %f", aggregated.AverageMI)
		}

		if aggregated.AverageMIwoc != 27.5 {
			t.Errorf("Expected 27.5, got %f", aggregated.AverageMIwoc)
		}

		if aggregated.AverageMIcw != 32.5 {
			t.Errorf("Expected 32.5, got %f", aggregated.AverageMIcw)
		}

		// Average per method
		if aggregated.AverageMIPerMethod != 22.5 {
			t.Errorf("Expected AverageMIPerMethod, got %f", aggregated.AverageMIPerMethod)
		}
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
