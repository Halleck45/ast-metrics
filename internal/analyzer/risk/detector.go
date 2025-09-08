package risk

import (
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// RiskItem represents a detected risk with a simple readable structure for reporting
type RiskItem struct {
	ID       string  // stable identifier
	Title    string  // short human-readable title
	Severity float64 // 0..1, higher is worse
	Details  string  // additional information
}

// Detector defines a simple risk detector interface
// A detector inspects a file (and optionally its classes) to produce zero or more RiskItem
// The logic must be simple and readable.
// Detectors should not mutate the file.
// Note: We intentionally do not change the protobuf schema; these risks are ephemeral for reporting.
type Detector interface {
	Name() string
	Detect(file *pb.File) []RiskItem
}
