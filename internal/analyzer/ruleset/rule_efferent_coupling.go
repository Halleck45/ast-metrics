package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type efferentCouplingRule struct {
	cfg *configuration.ConfigurationDefaultRule
}

func NewEfferentCouplingRule(c *configuration.ConfigurationDefaultRule) Rule {
	return &efferentCouplingRule{cfg: c}
}

func (r *efferentCouplingRule) Name() string { return "efferent_coupling" }

func (r *efferentCouplingRule) Description() string {
	return "Checks the efferent coupling of files/classes"
}

func (r *efferentCouplingRule) CheckFile(file *pb.File, addError func(string), addSuccess func(string)) {
	if r.cfg == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Coupling == nil {
		return
	}

	// Exclusions
	if r.cfg.ExcludePatterns != nil {
		for _, pattern := range r.cfg.ExcludePatterns {
			if regexp.MustCompile(pattern).MatchString(file.Path) {
				return
			}
		}
	}

	value := int(file.Stmts.Analyze.Coupling.Efferent)
	if r.cfg.Max > 0 && value > r.cfg.Max {
		addError(fmt.Sprintf("Efferent coupling too high in file %s: got %d (max: %d)", file.Path, value, r.cfg.Max))
		return
	}
	if r.cfg.Min > 0 && value < r.cfg.Min {
		addError(fmt.Sprintf("Efferent coupling too low in file %s: got %d (min: %d)", file.Path, value, r.cfg.Min))
		return
	}
	addSuccess(fmt.Sprintf("Efferent coupling OK in file %s", file.Path))
}
