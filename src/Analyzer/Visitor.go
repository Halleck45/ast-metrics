package Analyzer

import (
    pb "github.com/halleck45/ast-metrics/src/NodeType"
    "github.com/golang/protobuf/proto"
)

type Visitor interface {
    Visit(stmts *pb.Stmts, parents *pb.Stmts)
    LeaveNode(stmts *pb.Stmts)
}

type HelperRecursionVisitor struct {}
func (recurser *HelperRecursionVisitor) Recurse(stmts *pb.Stmts, visitor Visitor) {

    if stmts == nil {
        return
    }

    // initialize results
    if stmts.Analyze == nil {
        stmts.Analyze = &pb.Analyze{}
        stmts.Analyze.Complexity = &pb.Complexity{Cyclomatic: proto.Int32(1)}
    }

    // foreach type of statements
    for _, stmt := range stmts.StmtClass {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }
    for _, stmt := range stmts.StmtNamespace {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }
    for _, stmt := range stmts.StmtFunction {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }
    for _, stmt := range stmts.StmtTrait {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }
    for _, stmt := range stmts.StmtDecisionIf {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }
    for _, stmt := range stmts.StmtDecisionElseIf {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }
    for _, stmt := range stmts.StmtDecisionElse {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }
    for _, stmt := range stmts.StmtDecisionCase {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }
    for _, stmt := range stmts.StmtLoop {
        recurser.Recurse(stmt.Stmts, visitor)
        visitor.Visit(stmt.Stmts, stmts)
    }

    visitor.LeaveNode(stmts)
}