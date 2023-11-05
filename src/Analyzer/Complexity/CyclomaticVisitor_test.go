package Analyzer

import (
    "testing"
    "io/ioutil"
    pb "github.com/halleck45/ast-metrics/src/NodeType"
    "github.com/golang/protobuf/jsonpb"
)

func TestItCalculateCyclomaticComplexity(t *testing.T) {

   visitor := CyclomaticComplexityVisitor{}

    // Parsing existing file is simpler than creating a new one
    // We parse the file and then we calculate the complexity
    file := "testdata/complexity1.json"
    in, err := ioutil.ReadFile(file)
    if err != nil {
        t.Error("Failed to read file (" + file + "):", err)
    }
    pbFile := &pb.File{}
    if err := jsonpb.UnmarshalString(string(in), pbFile); err != nil {
        t.Error("Failed to parse file (" + file + "):", err)
    }

    ccn := visitor.Calculate(pbFile.Stmts.StmtClass[0].Stmts)

    // complexity should be 11
    if ccn != 11 {
        t.Error("Expected 11, got ", ccn)
    }
}