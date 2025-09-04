package ruleset

import (
	"fmt"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type maintainabilityRule struct {
	cfg *configuration.ConfigurationDefaultRule
}

func NewMaintainabilityRule(c *configuration.ConfigurationDefaultRule) Rule {
	return &maintainabilityRule{cfg: c}
}

func (r *maintainabilityRule) Name() string {
	return "maintainability"
}

func (r *maintainabilityRule) Description() string {
	return "Checks the maintainability of the code"
}

func (r *maintainabilityRule) CheckFile(file *pb.File, addError func(string), addSuccess func(string)) {
	if r.cfg == nil || file.Stmts == nil {
		return
	}
	// apply to classes
	hasAny := false
	for _, class := range file.Stmts.StmtClass {
		if class.Stmts == nil || class.Stmts.Analyze == nil || class.Stmts.Analyze.Maintainability == nil || class.Stmts.Analyze.Maintainability.MaintainabilityIndex == nil {
			continue
		}
		hasAny = true
		value := int(*class.Stmts.Analyze.Maintainability.MaintainabilityIndex)
		if r.cfg.Max > 0 && value > r.cfg.Max {
			addError(fmt.Sprintf("Maintainability too high in file %s: got %d (max: %d)", file.Path, value, r.cfg.Max))
			return
		}
		if r.cfg.Min > 0 && value < r.cfg.Min {
			addError(fmt.Sprintf("Maintainability too low in file %s: got %d (min: %d)", file.Path, value, r.cfg.Min))
			return
		}
	}
	if hasAny {
		addSuccess(fmt.Sprintf("Maintainability OK in file %s", file.Path))
	}
}
