package Cli

import (
    "testing"

    "github.com/halleck45/ast-metrics/src/Analyzer"
    pb "github.com/halleck45/ast-metrics/src/NodeType"
)

func TestNewScreenHome(t *testing.T) {
    isInteractive := true
    files := []*pb.File{}
    projectAggregated := Analyzer.ProjectAggregated{}

    screenHome := NewScreenHome(isInteractive, files, projectAggregated)

    if screenHome.isInteractive != isInteractive {
        t.Errorf("Expected isInteractive to be %v, but got %v", isInteractive, screenHome.isInteractive)
    }

    if len(screenHome.files) != len(files) {
        t.Errorf("Expected files to be %v, but got %v", files, screenHome.files)
    }
}

func TestGetModel(t *testing.T) {
    isInteractive := true
    files := []*pb.File{}
    projectAggregated := Analyzer.ProjectAggregated{}

    screenHome := NewScreenHome(isInteractive, files, projectAggregated)
    model := screenHome.GetModel()

    if len(model.files) != len(files) {
        t.Errorf("Expected files to be %v, but got %v", files, model.files)
    }
}

func TestInit(t *testing.T) {
    model := modelChoices{}
    cmd := model.Init()

    if cmd != nil {
        t.Errorf("Expected cmd to be nil, but got %v", cmd)
    }
}