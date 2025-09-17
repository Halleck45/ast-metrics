package ruleset

import (
	"fmt"
	"strings"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

// Rule: Context missing on exported I/O APIs

type ruleContextMissing struct{}

func (r *ruleContextMissing) Name() string {
	return "no_context_missing"
}
func (r *ruleContextMissing) Description() string {
	return "Exported functions that perform I/O/HTTP/DB should accept context.Context as first parameter"
}
func (r *ruleContextMissing) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if file == nil || file.Stmts == nil {
		return
	}

	flagged := 0
	checked := 0

	// helper to decide if an external dependency string indicates IO/HTTP/DB
	isIO := func(ns string) bool {
		ns = strings.TrimSpace(ns)
		if ns == "" {
			return false
		}
		ns = strings.ToLower(ns)
		return strings.HasPrefix(ns, "net/") || ns == "net" || strings.HasPrefix(ns, "net/http") || strings.HasPrefix(ns, "io") || strings.HasPrefix(ns, "os") || strings.HasPrefix(ns, "database/") || strings.HasPrefix(ns, "database") || strings.HasPrefix(ns, "syscall") || strings.HasPrefix(ns, "bufio") || strings.HasPrefix(ns, "crypto/tls")
	}

	// helper to check if a function appears to perform IO based on its externals
	funcUsesIO := func(f *pb.StmtFunction) bool {
		if f == nil {
			return false
		}
		// 1) names attached as externals
		for _, n := range f.Externals {
			if n == nil {
				continue
			}
			if isIO(n.GetQualified()) || isIO(n.GetShort()) || isIO(n.GetPackage()) {
				return true
			}
		}
		// 2) explicit external dependencies inside function stmts
		if f.Stmts != nil {
			for _, d := range f.Stmts.StmtExternalDependencies {
				if isIO(d.GetNamespace()) || isIO(d.GetClassName()) || isIO(d.GetFrom()) {
					return true
				}
			}
		}
		return false
	}

	// helper to check if first parameter type is context.Context
	hasContextFirstParam := func(f *pb.StmtFunction) bool {
		if f == nil || len(f.Parameters) == 0 {
			return false
		}
		p := f.Parameters[0]
		t := strings.TrimSpace(strings.ToLower(p.GetType()))
		if t == "" { // sometimes type might be empty if not captured
			return false
		}
		// Accept forms like context.Context, *context.Context, stdlib alias still contains context.context when qualified
		return strings.Contains(t, "context.context") || strings.HasSuffix(t, " context.context") || strings.HasSuffix(t, "*context.context")
	}

	// Check top-level functions in file
	for _, f := range file.Stmts.StmtFunction {
		if f == nil || f.Name == nil {
			continue
		}
		name := f.Name.Short
		if name == "" || name[0] < 'A' || name[0] > 'Z' { // exported only
			continue
		}
		if !funcUsesIO(f) {
			continue
		}
		checked++
		if !hasContextFirstParam(f) {
			flagged++
		}
	}
	// Also check functions inside namespaces and classes
	for _, ns := range file.Stmts.StmtNamespace {
		if ns == nil || ns.Stmts == nil {
			continue
		}
		for _, f := range ns.Stmts.StmtFunction {
			if f == nil || f.Name == nil {
				continue
			}
			name := f.Name.Short
			if name == "" || name[0] < 'A' || name[0] > 'Z' { // exported only
				continue
			}
			if !funcUsesIO(f) {
				continue
			}
			checked++
			if !hasContextFirstParam(f) {
				flagged++
			}
		}
		for _, cls := range ns.Stmts.StmtClass {
			if cls == nil || cls.Stmts == nil {
				continue
			}
			for _, f := range cls.Stmts.StmtFunction {
				if f == nil || f.Name == nil {
					continue
				}
				name := f.Name.Short
				if name == "" || name[0] < 'A' || name[0] > 'Z' { // exported only
					continue
				}
				if !funcUsesIO(f) {
					continue
				}
				checked++
				if !hasContextFirstParam(f) {
					flagged++
				}
			}
		}
	}

	if checked == 0 {
		addSuccess("No exported I/O-related functions found")
		return
	}
	if flagged > 0 {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Message:  fmt.Sprintf("Found %d exported I/O-related functions without context.Context as first parameter out of %d checked in %s", flagged, checked, file.Path),
			Code:     r.Name(),
		})
		return
	}
	addSuccess("Exported I/O APIs accept context")
}
