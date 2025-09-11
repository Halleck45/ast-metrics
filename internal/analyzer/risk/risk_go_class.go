package risk

import (
	"strings"

	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// TooManyGoClassesDetector flags Go files that contain too many classes/types in a single file
// This is a heuristic to encourage file cohesion in Go projects.
type TooManyGoClassesDetector struct{}

func (d *TooManyGoClassesDetector) Name() string { return "risk_go_class" }

func (d *TooManyGoClassesDetector) Detect(file *pb.File) []RiskItem {
	items := []RiskItem{}
	if file == nil || file.ProgrammingLanguage == "" {
		return items
	}
	if !strings.EqualFold(file.ProgrammingLanguage, "Go") {
		return items
	}
	classes := engine.GetClassesInFile(file)
	if len(classes) > 3 {
		items = append(items, RiskItem{
			ID:       d.Name(),
			Title:    "Too many types in a single Go file",
			Severity: 0.6 + float64(len(classes)-3)*0.05,
			Details:  "This Go file declares multiple types. Consider splitting to improve cohesion.",
		})
	}
	return items
}
