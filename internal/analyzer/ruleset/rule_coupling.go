package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
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
	if c.cfg == nil || file.Stmts == nil {
		return
	}

	// Aggregate dependencies from all levels (file, namespace, class, function)
	dependencies := engine.GetDependenciesInFile(file)
	if len(dependencies) == 0 {
		return
	}

	hasError := false
	for _, forbidden := range c.cfg.Forbidden {
		fromRegex := regexp.MustCompile("(?i)" + forbidden.From)
		if !fromRegex.MatchString(file.Path) {
			continue
		}
		toRegex := regexp.MustCompile("(?i)" + forbidden.To)
		for _, dependency := range dependencies {
			if toRegex.MatchString(dependency.ClassName) {
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
		addSuccess("Coupling OK")
	}
}
