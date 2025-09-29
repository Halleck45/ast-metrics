package ruleset

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/pb"
)

// Rule: Max number of files per package (excluding doc.go)

type ruleMaxFilesPerPackage struct {
	max int
}

func (r *ruleMaxFilesPerPackage) Name() string {
	return "max_files_per_package"
}
func (r *ruleMaxFilesPerPackage) Description() string {
	return "Limit number of source files per package (excluding doc.go)"
}
func (r *ruleMaxFilesPerPackage) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	if file == nil || file.Path == "" {
		return
	}
	dir := filepath.Dir(file.Path)
	pkg := ""
	if len(file.Stmts.StmtNamespace) > 0 && file.Stmts.StmtNamespace[0] != nil && file.Stmts.StmtNamespace[0].Name != nil {
		pkg = file.Stmts.StmtNamespace[0].Name.Short
	}
	count := 0
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			if path != dir {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		base := filepath.Base(path)
		if strings.EqualFold(base, "doc.go") {
			return nil
		}
		// ensure same package
		b, _ := os.ReadFile(path)
		if hasPackage(string(b), pkg) {
			count++
		}
		return nil
	})
	if r.max > 0 && count > r.max {
		addError(issue.RequirementError{
			Severity: issue.SeverityMedium,
			Message:  fmt.Sprintf("Package '%s' has %d files > %d in %s", pkg, count, r.max, dir),
			Code:     r.Name(),
		})
		return
	}
	addSuccess(fmt.Sprintf("Package '%s' has %d files ≤ %d in %s", pkg, count, r.max, dir))
}

func hasPackage(src, pkg string) bool {
	if pkg == "" {
		// if unknown, accept first seen
		re := regexp.MustCompile(`(?m)^\s*package\s+([a-zA-Z0-9_]+)\b`)
		m := re.FindStringSubmatch(src)
		return len(m) > 1
	}
	re := regexp.MustCompile(`(?m)^\s*package\s+` + regexp.QuoteMeta(pkg) + `\b`)
	return re.FindStringIndex(src) != nil
}
