package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type couplingRule struct {
	cfg *configuration.ConfigurationCouplingRule
}

func NewCouplingRule(c *configuration.ConfigurationCouplingRule) Rule {
	return &couplingRule{cfg: c}
}

func (c *couplingRule) Name() string {
	return "coupling"
}

func (c *couplingRule) Description() string {
	return "Checks for forbidden coupling between packages"
}

func (c *couplingRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if c.cfg == nil || file.Stmts == nil || file.Stmts.StmtExternalDependencies == nil {
		return
	}

	hasError := false
	for _, forbidden := range c.cfg.Forbidden {
		if !regexp.MustCompile(forbidden.From).MatchString(file.Path) {
			continue
		}
		for _, dependency := range file.Stmts.StmtExternalDependencies {
			if regexp.MustCompile(forbidden.To).MatchString(dependency.ClassName) {
				addError(issue.RequirementError{
					Severity: issue.SeverityUnknown,
					Code:     c.Name(),
					Message:  fmt.Sprintf("Forbidden coupling between %s and %s", file.Path, dependency.ClassName),
				})
				hasError = true
				break
			}
		}
	}

	if !hasError {
		addSuccess("Coupling OK in file " + file.Path)
	}
}

// Ensure imports used
var _ = configuration.ConfigurationDefaultRule{}
