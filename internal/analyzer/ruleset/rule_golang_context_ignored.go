package ruleset

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/halleck45/ast-metrics/internal/analyzer/issue"
	pb "github.com/halleck45/ast-metrics/internal/nodetype"
)

// Rule: Context ignored (use of context.Background/TODO when a context is available)

type ruleContextIgnored struct{}

func (r *ruleContextIgnored) Name() string { return "no_context_ignored" }
func (r *ruleContextIgnored) Description() string {
	return "Avoid ignoring context: prefer passing received context instead of context.Background/TODO"
}
func (r *ruleContextIgnored) CheckFile(file *pb.File, addError func(issue.RequirementError), addSuccess func(string)) {
	b, err := os.ReadFile(file.Path)
	if err != nil {
		return
	}
	src := string(b)
	reBg := regexp.MustCompile(`\bcontext\.(Background|TODO)\s*\(`)
	flagged := 0
	for _, fn := range file.Stmts.StmtFunction {
		if fn == nil || fn.Name == nil {
			continue
		}
		// if function has a parameter likely to be context and body uses context.Background/TODO -> flag
		hasCtxParam := false
		for _, p := range fn.Parameters {
			if p != nil && strings.EqualFold(p.GetName(), "ctx") {
				hasCtxParam = true
				break
			}
		}
		if !hasCtxParam {
			continue
		}
		// Extract function body snippet by searching for "func name(" and the following braces
		if containsInFunctionBody(src, fn.Name.Short, reBg) {
			flagged++
		}
	}
	if flagged > 0 {
		addError(issue.RequirementError{
			Code:     r.Name(),
			Severity: issue.SeverityLow,
			Message:  fmt.Sprintf("Avoid ignoring context: found %d use(s) of context.Background/TODO in functions receiving context", flagged),
		})
		return
	}
	addSuccess(fmt.Sprintf("[%s] No context misuse detected in %s", r.Name(), file.Path))
}

// containsInFunctionBody checks if pattern matches inside the body of a given function
func containsInFunctionBody(src, funcName string, re *regexp.Regexp) bool {
	reSig := regexp.MustCompile(`(?s)\bfunc\s+` + regexp.QuoteMeta(funcName) + `\s*\([^)]*\)\s*{`) // start of func
	loc := reSig.FindStringIndex(src)
	if loc == nil {
		return false
	}
	// naive brace matching from loc[1]
	depth := 0
	for i := loc[1]; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			if depth == 0 {
				return false
			}
			depth--
			if depth == 0 {
				segment := src[loc[1]:i]
				return re.FindStringIndex(segment) != nil
			}
		}
	}
	return false
}
