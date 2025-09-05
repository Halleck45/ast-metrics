package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type afferentCouplingRule struct {
	cfg *configuration.ConfigurationDefaultRule
}

func NewAfferentCouplingRule(c *configuration.ConfigurationDefaultRule) Rule {
	return &afferentCouplingRule{cfg: c}
}

func (r *afferentCouplingRule) Name() string { return "afferent_coupling" }

func (r *afferentCouplingRule) Description() string {
	return "Checks the afferent coupling of files/classes"
}

func (r *afferentCouplingRule) CheckFile(file *pb.File, addError func(string), addSuccess func(string)) {
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

	value := int(file.Stmts.Analyze.Coupling.Afferent)
	if r.cfg.Max > 0 && value > r.cfg.Max {
		addError(fmt.Sprintf("Afferent coupling too high in file %s: got %d (max: %d)", file.Path, value, r.cfg.Max))
		return
	}
	if r.cfg.Min > 0 && value < r.cfg.Min {
		addError(fmt.Sprintf("Afferent coupling too low in file %s: got %d (min: %d)", file.Path, value, r.cfg.Min))
		return
	}
	addSuccess(fmt.Sprintf("Afferent coupling OK in file %s", file.Path))
}
