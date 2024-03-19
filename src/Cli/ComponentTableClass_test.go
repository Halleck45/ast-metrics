package Cli

import (
	"testing"

	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/stretchr/testify/assert"
)

func TestNewComponentTableClass(t *testing.T) {
	isInteractive := true
	files := []*pb.File{
		{
			ProgrammingLanguage: "Go",
		},
		{
			ProgrammingLanguage: "Python",
		},
	}

	table := NewComponentTableClass(isInteractive, files)

	if table.isInteractive != isInteractive {
		t.Errorf("Expected isInteractive to be %v, got %v", isInteractive, table.isInteractive)
	}

	if len(table.files) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(table.files))
	}

	for i, file := range table.files {
		if file.ProgrammingLanguage != files[i].ProgrammingLanguage {
			t.Errorf("Expected file at index %d to have ProgrammingLanguage %s, got %s", i, files[i].ProgrammingLanguage, file.ProgrammingLanguage)
		}
	}

	if table.sortColumnIndex != 0 {
		t.Errorf("Expected sortColumnIndex to be 0, got %d", table.sortColumnIndex)
	}
}

func TestComponentTableClass_Render(t *testing.T) {

	mi := float32(120)
	ccn := int32(5)
	loc := int32(100)
	halsteadLength := int32(100)
	halsteadVolume := float32(100)

	files := []*pb.File{
		{
			Path: "file1.php",
			Stmts: &pb.Stmts{
				StmtClass: []*pb.StmtClass{
					{
						Name: &pb.Name{
							Qualified: "ClassA",
							Short:     "ClassA",
						},
						Stmts: &pb.Stmts{
							Analyze: &pb.Analyze{
								Complexity: &pb.Complexity{
									Cyclomatic: &ccn,
								},
								Maintainability: &pb.Maintainability{
									MaintainabilityIndex: &mi,
								},
								Volume: &pb.Volume{
									Loc:            &loc,
									HalsteadLength: &halsteadLength,
									HalsteadVolume: &halsteadVolume,
								},
							},
						},
					},
				},
			},
		},
	}
	component := NewComponentTableClass(false, files)
	component.Init()

	rendered := component.Render()

	assert.Contains(t, rendered, "Use arrows to navigate and esc to quit", "Help is present")
	assert.Contains(t, rendered, "ClassA", "ClassA is present")
}

func TestComponentTableClass_Sort(t *testing.T) {

	// class 1
	mi1 := float32(120)
	ccn1 := int32(5)
	loc1 := int32(100)
	halsteadLength1 := int32(5)
	halsteadVolume := float32(7)

	//  class 2
	mi2 := float32(110)
	ccn2 := int32(10)
	loc2 := int32(80)
	halsteadLength2 := int32(7)

	files := []*pb.File{
		{
			Path: "file1.php",
			Stmts: &pb.Stmts{
				StmtClass: []*pb.StmtClass{
					{
						Name: &pb.Name{
							Qualified: "ClassA",
							Short:     "ClassA",
						},
						Stmts: &pb.Stmts{
							Analyze: &pb.Analyze{
								Complexity: &pb.Complexity{
									Cyclomatic: &ccn1,
								},
								Maintainability: &pb.Maintainability{
									MaintainabilityIndex: &mi1,
								},
								Volume: &pb.Volume{
									Loc:            &loc1,
									HalsteadLength: &halsteadLength1,
									HalsteadVolume: &halsteadVolume,
								},
							},
						},
					},
					{
						Name: &pb.Name{
							Qualified: "ClassB",
							Short:     "ClassB",
						},
						Stmts: &pb.Stmts{
							Analyze: &pb.Analyze{
								Complexity: &pb.Complexity{
									Cyclomatic: &ccn2,
								},
								Maintainability: &pb.Maintainability{
									MaintainabilityIndex: &mi2,
								},
								Volume: &pb.Volume{
									Loc:            &loc2,
									HalsteadLength: &halsteadLength2,
									HalsteadVolume: &halsteadVolume,
								},
							},
						},
					},
				},
			},
		},
	}
	component := NewComponentTableClass(false, files)
	component.Init()

	// Sort by maintainability index
	component.SortByMaintainabilityIndex()
	firstRow := component.table.Rows()[0]
	assert.Contains(t, firstRow[0], "ClassB", "ClassB is first")

	// Sort by name
	component.SortByName()
	firstRow = component.table.Rows()[0]
	assert.Contains(t, firstRow[0], "ClassA", "ClassA is first")

	// Sort by cyclomatic complexity
	component.SortByCyclomaticComplexity()
	firstRow = component.table.Rows()[0]
	assert.Contains(t, firstRow[0], "ClassB", "ClassB is first")

	// Sort by number of methods
	component.SortByNumberOfMethods()
	firstRow = component.table.Rows()[0]
	assert.Contains(t, firstRow[0], "ClassB", "ClassB is first")

	// Sort again by name
	component.SortByName()
	firstRow = component.table.Rows()[0]
	assert.Contains(t, firstRow[0], "ClassA", "ClassA is first")

}
