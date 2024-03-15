package Cli

import (
    "testing"
    "github.com/halleck45/ast-metrics/src/Analyzer"
    pb "github.com/halleck45/ast-metrics/src/NodeType"
)

func TestScreenTableClassNewScreenTableClass(t *testing.T) {
    files := []*pb.File{}
    projectAggregated := Analyzer.ProjectAggregated{}
    screenTableClass := NewScreenTableClass(true, files, projectAggregated)

    if screenTableClass.isInteractive != true {
        t.Errorf("Expected isInteractive to be true, got %v", screenTableClass.isInteractive)
    }

    if len(screenTableClass.files) != len(files) {
        t.Errorf("Expected files to be %v, got %v", files, screenTableClass.files)
    }
}

func TestScreenTableClassGetScreenName(t *testing.T) {
    files := []*pb.File{}
    projectAggregated := Analyzer.ProjectAggregated{}
    screenTableClass := NewScreenTableClass(true, files, projectAggregated)

    screenName := screenTableClass.GetScreenName()

    if screenName != "Classes and object oriented statistics" {
        t.Errorf("Expected screen name to be 'Classes and object oriented statistics', got %v", screenName)
    }
}

func TestScreenTableClassGetModel(t *testing.T) {
    files := []*pb.File{}
    projectAggregated := Analyzer.ProjectAggregated{}
    screenTableClass := NewScreenTableClass(true, files, projectAggregated)

    model := screenTableClass.GetModel()

    if model == nil {
        t.Errorf("Expected model to not be nil")
    }
}