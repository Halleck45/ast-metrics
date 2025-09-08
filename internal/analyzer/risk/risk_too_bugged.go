package risk

import (
	"fmt"

	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// TooBuggedDetector flags files or classes when Halstead estimated bugs are high.
type TooBuggedDetector struct{}

func (d *TooBuggedDetector) Name() string { return "risk_too_bugged" }

func (d *TooBuggedDetector) Detect(file *pb.File) []RiskItem {
	items := []RiskItem{}
	if file == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Volume == nil || file.Stmts.Analyze.Volume.HalsteadBugs == nil {
		return items
	}
	bugs := *file.Stmts.Analyze.Volume.HalsteadBugs
	if bugs >= 0.5 { // arbitrary but pragmatic threshold
		items = append(items, RiskItem{
			ID:       d.Name(),
			Title:    "High estimated bugs (Halstead)",
			Severity: clamp01Float(bugs/2.0 + 0.3),
			Details:  fmt.Sprintf("Estimated bugs: %.2f", bugs),
		})
	}
	return items
}

func clamp01Float(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
