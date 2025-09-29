package analyzer

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/pb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHalsteadMetricsVisitor(t *testing.T) {

	// from https://en.wikipedia.org/wiki/Halstead_complexity_measures
	json := `
{
    "path": "file.php",
    "stmts": {
        "stmtFunction": [
            {
                "name": {
                    "short": "main",
                    "qualified": "\\main"
                },
                "stmts": {},
                "location": {
                    "startLine": 3,
                    "startFilePos": 7,
                    "endLine": 7,
                    "endFilePos": 85
                },
                "operators": [
                    {
                        "name": "Assign"
                    },
                    {
                        "name": "BinaryOp_Div"
                    },
                    {
                        "name": "FuncCall"
                    }
                ],
                "operands": [
                    {
                        "name": "avg"
                    }
                ],
                "parameters": [
                    {
                        "name": "a",
                        "type": ""
                    },
                    {
                        "name": "b",
                        "type": ""
                    },
                    {
                        "name": "c",
                        "type": ""
                    }
                ],
                "linesOfCode": {
                    "linesOfCode": 5,
                    "logicalLinesOfCode": 5
                }
            }
        ]
    },
    "linesOfCode": {
        "linesOfCode": 5,
        "logicalLinesOfCode": 5
    }
}
`

	pbFile := &pb.File{}
	if err := protojson.Unmarshal([]byte(json), pbFile); err != nil {
		panic(err)
	}

	visitor := HalsteadMetricsVisitor{}

	visitor.Visit(pbFile.Stmts, pbFile.Stmts)
	visitor.LeaveNode(pbFile.Stmts)

	// Add your assertions here based on the expected values of the Halstead metrics
	// Check the Halstead metrics for the function with the "add" and "subtract" operators
	if *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadVocabulary != int32(4) {
		t.Errorf("Expected 4, got %d", *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadVocabulary)
	}

	if *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadLength != int32(4) {
		t.Errorf("Expected 4, got %d", *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadLength)
	}

	if *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadEstimatedLength != float64(4.754887502163469) {
		t.Errorf("Expected 4.754887502163469, got %f", *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadEstimatedLength)
	}

	if *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadVolume != float64(8) {
		t.Errorf("Expected 8, got %f", *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadVolume)
	}

	if *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadDifficulty != float64(1.5) {
		t.Errorf("Expected 1.5, got %f", *pbFile.Stmts.StmtFunction[0].Stmts.Analyze.Volume.HalsteadDifficulty)
	}
}

func TestHalsteadMetricsVisitor_LeaveNode(t *testing.T) {
	visitor := HalsteadMetricsVisitor{}

	stmts := &pb.Stmts{
		StmtClass: []*pb.StmtClass{
			{
				Stmts: &pb.Stmts{
					StmtFunction: []*pb.StmtFunction{
						{
							Stmts: &pb.Stmts{
								Analyze: &pb.Analyze{
									Volume: &pb.Volume{
										HalsteadVocabulary:      proto.Int32(2),
										HalsteadLength:          proto.Int32(2),
										HalsteadEstimatedLength: proto.Float64(2.5),
										HalsteadVolume:          proto.Float64(2.5),
										HalsteadDifficulty:      proto.Float64(2.5),
										HalsteadEffort:          proto.Float64(2.5),
										HalsteadTime:            proto.Float64(2.5),
									},
								},
							},
						},
						{
							Stmts: &pb.Stmts{
								Analyze: &pb.Analyze{
									Volume: &pb.Volume{
										HalsteadVocabulary:      proto.Int32(4),
										HalsteadLength:          proto.Int32(4),
										HalsteadEstimatedLength: proto.Float64(4.5),
										HalsteadVolume:          proto.Float64(4.5),
										HalsteadDifficulty:      proto.Float64(4.5),
										HalsteadEffort:          proto.Float64(4.5),
										HalsteadTime:            proto.Float64(4.5),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	visitor.LeaveNode(stmts)

	if *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadVocabulary != int32(3) {
		t.Errorf("Expected 3, got %d", *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadVocabulary)
	}

	if *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadLength != int32(3) {
		t.Errorf("Expected 3, got %d", *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadLength)
	}

	if *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadEstimatedLength != float64(3.5) {
		t.Errorf("Expected 3.5, got %f", *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadEstimatedLength)
	}

	if *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadVolume != float64(3.5) {
		t.Errorf("Expected 3.5, got %f", *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadVolume)
	}

	if *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadDifficulty != float64(3.5) {
		t.Errorf("Expected 3.5, got %f", *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadDifficulty)
	}

	if *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadEffort != float64(3.5) {
		t.Errorf("Expected 3.5, got %f", *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadEffort)
	}

	if *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadTime != float64(3.5) {
		t.Errorf("Expected 3.5, got %f", *stmts.StmtClass[0].Stmts.Analyze.Volume.HalsteadTime)
	}
}
