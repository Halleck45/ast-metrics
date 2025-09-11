package risk

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// TooManyResponsibilityDetector flags classes with too many methods or poor cohesion (LCOM4 > 1)
type TooManyResponsibilityDetector struct{}

func (d *TooManyResponsibilityDetector) Name() string { return "risk_too_many_responsability" }

func (d *TooManyResponsibilityDetector) Detect(file *pb.File) []RiskItem {
	items := []RiskItem{}
	if file == nil {
		return items
	}
	for _, cls := range engine.GetClassesInFile(file) {
		var methods int
		if cls.Stmts != nil {
			methods = len(cls.Stmts.StmtFunction)
		}
		var lcom4 int32
		if cls.Stmts != nil && cls.Stmts.Analyze != nil && cls.Stmts.Analyze.ClassCohesion != nil && cls.Stmts.Analyze.ClassCohesion.Lcom4 != nil {
			lcom4 = *cls.Stmts.Analyze.ClassCohesion.Lcom4
		}
		if methods >= 20 || lcom4 > 1 {
			sev := 0.5
			if methods >= 30 {
				sev = 0.8
			}
			if lcom4 > 1 && sev < 0.7 {
				sev = 0.7
			}
			title := "Class may have too many responsibilities"
			details := fmt.Sprintf("Class %s has %d methods and LCOM4=%d", cls.Name.GetQualified(), methods, lcom4)
			items = append(items, RiskItem{ID: d.Name(), Title: title, Severity: sev, Details: details})
		}
	}
	return items
}
