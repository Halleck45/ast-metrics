package risk

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// TooManyEfferentCouplingDetector flags classes with high efferent coupling
// Uses simple thresholds to keep code readable

type TooManyEfferentCouplingDetector struct{}

func (d *TooManyEfferentCouplingDetector) Name() string { return "risk_too_many_efferent_coupling" }

func (d *TooManyEfferentCouplingDetector) Detect(file *pb.File) []RiskItem {
	items := []RiskItem{}
	if file == nil {
		return items
	}
	for _, cls := range engine.GetClassesInFile(file) {
		if cls.Stmts == nil || cls.Stmts.Analyze == nil || cls.Stmts.Analyze.Coupling == nil {
			continue
		}
		ef := cls.Stmts.Analyze.Coupling.Efferent
		if ef >= 20 {
			sev := 0.6
			if ef >= 40 {
				sev = 0.85
			}
			items = append(items, RiskItem{
				ID:       d.Name(),
				Title:    "Excessive efferent coupling",
				Severity: sev,
				Details:  fmt.Sprintf("Class %s depends on %d other classes", cls.Name.GetQualified(), ef),
			})
		}
	}
	return items
}
