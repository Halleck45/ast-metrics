package ruleset

import (
	"fmt"
	"strings"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// Rule: No package name in function/method names
// Do not include the package name in exported function or method identifiers
// Severity: Low→Medium

type ruleNoPkgInMethod struct{}

func (r *ruleNoPkgInMethod) Name() string {
	return "no_package_name_in_method"
}
func (r *ruleNoPkgInMethod) Description() string {
	return "Do not include the package name in exported function or method identifiers"
}
func (r *ruleNoPkgInMethod) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if file == nil || file.Stmts == nil {
		return
	}

	if file.ProgrammingLanguage != "Golang" {
		return
	}

	pkg := ""
	if len(file.Stmts.StmtNamespace) > 0 && file.Stmts.StmtNamespace[0] != nil && file.Stmts.StmtNamespace[0].Name != nil {
		pkg = file.Stmts.StmtNamespace[0].Name.Short
	}
	if pkg == "" {
		return
	}
	lowPkg := strings.ToLower(pkg)
	flagged := 0
	check := func(fn *pb.StmtFunction) {
		if fn == nil || fn.Name == nil {
			return
		}
		name := fn.Name.Short
		if name == "" {
			return
		}
		// only consider exported identifiers (starting with uppercase) to reduce false positives
		if name[0] < 'A' || name[0] > 'Z' {
			return
		}
		if strings.Contains(strings.ToLower(name), lowPkg) {
			flagged++
			addError(issue.RequirementError{
				Severity: issue.SeverityMedium,
				Message:  fmt.Sprintf("Function/method name '%s()' contains package name '%s'", name, pkg),
				Code:     r.Name(),
			})
		}
	}
	for _, f := range file.Stmts.StmtFunction {
		check(f)
	}
	for _, ns := range file.Stmts.StmtNamespace {
		if ns == nil || ns.Stmts == nil {
			continue
		}
		for _, f := range ns.Stmts.StmtFunction {
			check(f)
		}
		for _, c := range ns.Stmts.StmtClass { // methods on types
			if c == nil || c.Stmts == nil {
				continue
			}
			for _, f := range c.Stmts.StmtFunction {
				check(f)
			}
		}
	}
	if flagged == 0 {
		addSuccess("Names of functions/methods does not contain package name OK")
	}
}
