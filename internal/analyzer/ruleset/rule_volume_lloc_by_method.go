package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type llocByMethodRule struct {
	Cfg *configuration.ConfigurationDefaultRule
}

func NewLlocByMethodRule(c *configuration.ConfigurationDefaultRule) Rule {
	return &llocByMethodRule{Cfg: c}
}

func (r *llocByMethodRule) Name() string { return "lloc_by_method" }

func (r *llocByMethodRule) Description() string {
	return "Checks the logical lines of code by method/function"
}

func (r *llocByMethodRule) CheckFile(file *pb.File, addError func(string), addSuccess func(string)) {
	if r.Cfg == nil || file.Stmts == nil || file.Stmts.StmtFunction == nil {
		return
	}

	// Exclusions
	if r.Cfg.ExcludePatterns != nil {
		for _, pattern := range r.Cfg.ExcludePatterns {
			if regexp.MustCompile(pattern).MatchString(file.Path) {
				return
			}
		}
	}

	ok := true
	for _, f := range file.Stmts.StmtFunction {
		if f == nil || f.LinesOfCode == nil || f.LinesOfCode.LogicalLinesOfCode == 0 {
			continue
		}
		value := int(f.LinesOfCode.LogicalLinesOfCode)
		if r.Cfg.Max > 0 && value > r.Cfg.Max {
			addError(fmt.Sprintf("LLOC too high in method %s (file %s): got %d (max: %d)", f.Name.Short, file.Path, value, r.Cfg.Max))
			ok = false
			continue
		}
		if r.Cfg.Min > 0 && value < r.Cfg.Min {
			addError(fmt.Sprintf("LLOC too low in method %s (file %s): got %d (min: %d)", f.Name.Short, file.Path, value, r.Cfg.Min))
			ok = false
			continue
		}
	}
	if ok {
		addSuccess(fmt.Sprintf("LLOC by method OK in file %s", file.Path))
	}
}
