package Cli

import (
	pb "github.com/halleck45/ast-metrics/src/NodeType"
	"github.com/halleck45/ast-metrics/src/Analyzer"
)

type Renderer interface {
    Render (pbFiles []pb.File, aggregated Analyzer.Aggregated)
}