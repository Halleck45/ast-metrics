package Analyzer

import (
	"os"
	"testing"

	"github.com/halleck45/ast-metrics/src/Engine/Golang"
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestItCalculateCyclomaticComplexity(t *testing.T) {

	visitor := CyclomaticComplexityVisitor{}

	json := `
    {
        "path": "\/var\/www\/ast-metrics\/src\/Php\/phpsources\/tests\/resources\/file2.php",
        "stmts": {
            "stmtClass": [
                {
                    "name": {
                        "short": "Foo",
                        "qualified": "\\Foo"
                    },
                    "stmts": {
                        "stmtFunction": [
                            {
                                "name": {
                                    "short": "example",
                                    "qualified": "\\example"
                                },
                                "stmts": {
                                    "stmtDecisionIf": [
                                        {
                                            "stmts": {
                                                "stmtDecisionIf": [
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 6,
                                                            "startFilePos": 127,
                                                            "endLine": 12,
                                                            "endFilePos": 292
                                                        }
                                                    }
                                                ],
                                                "stmtDecisionElseIf": [
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 8,
                                                            "startFilePos": 185,
                                                            "endLine": 10,
                                                            "endFilePos": 245
                                                        }
                                                    }
                                                ],
                                                "stmtDecisionElse": [
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 10,
                                                            "startFilePos": 247,
                                                            "endLine": 12,
                                                            "endFilePos": 292
                                                        }
                                                    }
                                                ]
                                            },
                                            "location": {
                                                "startLine": 5,
                                                "startFilePos": 99,
                                                "endLine": 36,
                                                "endFilePos": 897
                                            }
                                        }
                                    ],
                                    "stmtDecisionElseIf": [
                                        {
                                            "stmts": {
                                                "stmtLoop": [
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 14,
                                                            "startFilePos": 336,
                                                            "endLine": 16,
                                                            "endFilePos": 393
                                                        }
                                                    }
                                                ]
                                            },
                                            "location": {
                                                "startLine": 13,
                                                "startFilePos": 304,
                                                "endLine": 17,
                                                "endFilePos": 403
                                            }
                                        },
                                        {
                                            "stmts": {
                                                "stmtLoop": [
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 18,
                                                            "startFilePos": 437,
                                                            "endLine": 20,
                                                            "endFilePos": 505
                                                        }
                                                    }
                                                ]
                                            },
                                            "location": {
                                                "startLine": 17,
                                                "startFilePos": 405,
                                                "endLine": 21,
                                                "endFilePos": 515
                                            }
                                        }
                                    ],
                                    "stmtDecisionElse": [
                                        {
                                            "stmts": {
                                                "stmtDecisionCase": [
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 23,
                                                            "startFilePos": 566,
                                                            "endLine": 25,
                                                            "endFilePos": 629
                                                        }
                                                    },
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 26,
                                                            "startFilePos": 647,
                                                            "endLine": 28,
                                                            "endFilePos": 710
                                                        }
                                                    },
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 29,
                                                            "startFilePos": 728,
                                                            "endLine": 31,
                                                            "endFilePos": 791
                                                        }
                                                    },
                                                    {
                                                        "stmts": {},
                                                        "location": {
                                                            "startLine": 32,
                                                            "startFilePos": 809,
                                                            "endLine": 34,
                                                            "endFilePos": 873
                                                        }
                                                    }
                                                ],
                                                "stmtDecisionSwitch": [
                                                    {
                                                        "location": {
                                                            "startLine": 22,
                                                            "startFilePos": 536,
                                                            "endLine": 35,
                                                            "endFilePos": 887
                                                        }
                                                    }
                                                ]
                                            },
                                            "location": {
                                                "startLine": 21,
                                                "startFilePos": 517,
                                                "endLine": 36,
                                                "endFilePos": 897
                                            }
                                        }
                                    ]
                                },
                                "location": {
                                    "startLine": 4,
                                    "startFilePos": 63,
                                    "endLine": 37,
                                    "endFilePos": 903
                                },
                                "operators": [
                                    {
                                        "name": "Expr_FuncCall"
                                    },
                                    {
                                        "name": "Expr_FuncCall"
                                    },
                                    {
                                        "name": "Expr_FuncCall"
                                    },
                                    {
                                        "name": "Expr_FuncCall"
                                    },
                                    {
                                        "name": "Expr_FuncCall"
                                    },
                                    {
                                        "name": "Expr_FuncCall"
                                    },
                                    {
                                        "name": "Expr_FuncCall"
                                    },
                                    {
                                        "name": "Expr_FuncCall"
                                    },
                                    {
                                        "name": "Expr_FuncCall"
                                    }
                                ]
                            }
                        ]
                    },
                    "location": {
                        "startLine": 3,
                        "startFilePos": 47,
                        "endLine": 38,
                        "endFilePos": 905
                    },
                    "comments": [
                        {
                            "location": {
                                "startLine": 2,
                                "startFilePos": 6,
                                "endLine": 2,
                                "endFilePos": 45
                            }
                        },
                        {
                            "location": {
                                "startLine": 2,
                                "startFilePos": 6,
                                "endLine": 2,
                                "endFilePos": 45
                            }
                        }
                    ]
                }
            ]
        }
    }
    `
	pbFile := &pb.File{}
	if err := protojson.Unmarshal([]byte(json), pbFile); err != nil {
		panic(err)
	}

	ccn := visitor.Calculate(pbFile.Stmts.StmtClass[0].Stmts)

	// complexity should be 11
	if ccn != 11 {
		t.Error("Expected 11, got ", ccn)
	}
}

func TestItCalculateCyclomaticComplexityForNotObjectOrientedLanguages(t *testing.T) {

	visitor := CyclomaticComplexityVisitor{}

	fileContent := `
    package main

    import "fmt"

    func example() {
        if true {
            if true {
                fmt.Println("Hello")
            }
        } else if true {
            fmt.Println("Hello")
        } else {
            fmt.Println("Hello")
        }
    }
    `

	// Create a temporary file
	tmpFile := t.TempDir() + "/test.php"
	if _, err := os.Create(tmpFile); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(tmpFile, []byte(fileContent), 0644); err != nil {
		t.Error(err)
	}

	pbFile := Golang.ParseGoFile(tmpFile)

	ccn := visitor.Calculate(pbFile.Stmts)

	if ccn != 3 {
		t.Error("Expected 3, got ", ccn)
	}
}
