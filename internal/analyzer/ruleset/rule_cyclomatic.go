package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type cyclomaticRule struct {
	cfg *configuration.ConfigurationDefaultRule
}

func NewCyclomaticRule(c *configuration.ConfigurationDefaultRule) Rule {
	return &cyclomaticRule{cfg: c}
}

func (r *cyclomaticRule) Name() string {
	return "cyclomatic_complexity"
}

func (r *cyclomaticRule) Description() string {
	return "Checks the cyclomatic complexity of functions"
}

func (r *cyclomaticRule) CheckFile(file *pb.File, addError func(string), addSuccess func(string)) {
	if r.cfg == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Complexity == nil || file.Stmts.Analyze.Complexity.Cyclomatic == nil {
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

	value := int(*file.Stmts.Analyze.Complexity.Cyclomatic)
	if r.cfg.Max > 0 && value > r.cfg.Max {
		addError(fmt.Sprintf("Cyclomatic complexity too high in file %s: got %d (max: %d)", file.Path, value, r.cfg.Max))
		return
	}
	if r.cfg.Min > 0 && value < r.cfg.Min {
		addError(fmt.Sprintf("Cyclomatic complexity too low in file %s: got %d (min: %d)", file.Path, value, r.cfg.Min))
		return
	}
	addSuccess(fmt.Sprintf("Cyclomatic complexity OK in file %s", file.Path))
}
