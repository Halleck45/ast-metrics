package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type locByMethodRule struct {
	Cfg *configuration.ConfigurationDefaultRule
}

func NewLocByMethodRule(c *configuration.ConfigurationDefaultRule) Rule {
	return &locByMethodRule{Cfg: c}
}

func (r *locByMethodRule) Name() string { return "loc_by_method" }

func (r *locByMethodRule) Description() string {
	return "Checks the lines of code by method/function"
}

func (r *locByMethodRule) CheckFile(file *pb.File, addError func(string), addSuccess func(string)) {
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
		if f == nil || f.LinesOfCode == nil || f.LinesOfCode.LinesOfCode == 0 {
			continue
		}
		value := int(f.LinesOfCode.LinesOfCode)
		if r.Cfg.Max > 0 && value > r.Cfg.Max {
			addError(fmt.Sprintf("LOC too high in method %s (file %s): got %d (max: %d)", f.Name.Short, file.Path, value, r.Cfg.Max))
			ok = false
			continue
		}
		if r.Cfg.Min > 0 && value < r.Cfg.Min {
			addError(fmt.Sprintf("LOC too low in method %s (file %s): got %d (min: %d)", f.Name.Short, file.Path, value, r.Cfg.Min))
			ok = false
			continue
		}
	}
	if ok {
		addSuccess(fmt.Sprintf("LOC by method OK in file %s", file.Path))
	}
}
