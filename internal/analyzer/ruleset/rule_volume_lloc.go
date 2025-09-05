package ruleset

import (
	"fmt"
	"regexp"

	"github.com/halleck45/ast-metrics/internal/configuration"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

type llocRule struct {
	Cfg *configuration.ConfigurationDefaultRule
}

func NewLlocRule(c *configuration.ConfigurationDefaultRule) Rule {
	return &llocRule{Cfg: c}
}

func (l *llocRule) Name() string {
	return "lloc"
}

func (l *llocRule) Description() string {
	return "Checks the logical lines of code in a file"
}

func (l *llocRule) CheckFile(file *pb.File, addError func(string), addSuccess func(string)) {
	if l.Cfg == nil || file.Stmts == nil || file.Stmts.Analyze == nil || file.Stmts.Analyze.Volume == nil || file.Stmts.Analyze.Volume.Lloc == nil {
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

	value := int(*file.Stmts.Analyze.Volume.Lloc)
	if l.Cfg.Max > 0 && value > l.Cfg.Max {
		addError(fmt.Sprintf("Logical lines of code too high in file %s: got %d (max: %d)", file.Path, value, l.Cfg.Max))
		return
	}
	if l.Cfg.Min > 0 && value < l.Cfg.Min {
		addError(fmt.Sprintf("Logical lines of code too low in file %s: got %d (min: %d)", file.Path, value, l.Cfg.Min))
		return
	}
	addSuccess(fmt.Sprintf("Logical lines of code OK in file %s", file.Path))
}
