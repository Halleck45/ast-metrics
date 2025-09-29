package ruleset

import (
	"fmt"
	"strings"
	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

type maxPublicMethodsRule struct {
	threshold int
}

func NewMaxPublicMethodsRule(threshold *int) Rule {
	if threshold == nil {
		return &maxPublicMethodsRule{threshold: 0}
	}
	return &maxPublicMethodsRule{threshold: *threshold}
}

func (r *maxPublicMethodsRule) Name() string {
	return "max_public_methods"
}

func (r *maxPublicMethodsRule) Description() string {
	return "Maximum number of public methods per class"
}

func (r *maxPublicMethodsRule) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if r.threshold == 0 || file.Stmts == nil || file.Stmts.StmtClass == nil {
		return
	}

	for _, class := range file.Stmts.StmtClass {
		if class.Stmts != nil && class.Stmts.StmtFunction != nil {
			publicCount := 0
			for _, method := range class.Stmts.StmtFunction {
				if method.Name != nil && method.Name.Short != "" {
					// Simple heuristic: method is public if it starts with uppercase (Go) or doesn't start with _ (other languages)
					name := method.Name.Short
					if len(name) > 0 && (name[0] >= 'A' && name[0] <= 'Z' || !strings.HasPrefix(name, "_")) {
						publicCount++
					}
				}
			}
			if publicCount > r.threshold {
				addError(issue.RequirementError{
					Severity: issue.SeverityMedium,
					Message:  fmt.Sprintf("Class has %d public methods, maximum allowed is %d", publicCount, r.threshold),
					Code:     r.Name(),
				})
				return
			}
		}
	}

	addSuccess(fmt.Sprintf("Max public methods OK in file %s", file.Path))
}
