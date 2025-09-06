package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type locRule struct {
	Cfg *configuration.ConfigurationDefaultRule
}

func NewLocRule(c *configuration.ConfigurationDefaultRule) Rule {
	return &locRule{Cfg: c}
}

func (l *locRule) Name() string {
	return "loc"
}

func (l *locRule) Description() string {
	return "Checks the lines of code in a file"
}

func (l *locRule) CheckFile(file *pb.File, addError func(string), addSuccess func(string)) {

	if l.Cfg == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Volume == nil || file.Stmts.Analyze.Volume.Loc == nil {
		return
	}
	// Exclusions
	if l.Cfg.ExcludePatterns != nil {
		for _, pattern := range l.Cfg.ExcludePatterns {
			if regexp.MustCompile(pattern).MatchString(file.Path) {
				return
			}
		}
	}

	value := int(*file.Stmts.Analyze.Volume.Loc)

	if l.Cfg.Max > 0 && value > l.Cfg.Max {
		addError(fmt.Sprintf("Lines of code too high in file %s: got %d (max: %d)", file.Path, value, l.Cfg.Max))
		return
	}
	if l.Cfg.Min > 0 && value < l.Cfg.Min {
		addError(fmt.Sprintf("Lines of code too low in file %s: got %d (min: %d)", file.Path, value, l.Cfg.Min))
		return
	}
	addSuccess(fmt.Sprintf("Lines of code OK in file %s", file.Path))
}
