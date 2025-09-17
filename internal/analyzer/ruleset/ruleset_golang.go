package ruleset

import "github.com/halleck45/ast-metrics/internal/configuration"

// golangRuleset defines Golang-specific best-practice rules
// This ruleset is opt-in and disabled by default; enable via requirements.rules.golang.enabled: true
// Thresholds are conservative defaults and can be tuned later via configuration if needed.
type golangRuleset struct {
	cfg *configuration.ConfigurationRequirements
}

func (g *golangRuleset) Category() string {
	return "golang"
}
func (g *golangRuleset) Description() string {
	return "Golang-specific best practices and API hygiene"
}
func (g *golangRuleset) Enabled() []Rule {
	// Return only rules enabled per configuration
	var out []Rule
	if g == nil || g.cfg == nil || g.cfg.Rules == nil || g.cfg.Rules.Golang == nil {
		return out
	}
	cfg := g.cfg.Rules.Golang
	isTrue := func(p *bool) bool { return p != nil && *p }
	if isTrue(cfg.NoPackageNameInMethod) {
		out = append(out, &ruleNoPkgInMethod{})
	}
	if cfg.MaxNesting != nil && *cfg.MaxNesting > 0 {
		out = append(out, &ruleMaxNesting{max: *cfg.MaxNesting})
	}
	if cfg.MaxFileSize != nil && *cfg.MaxFileSize > 0 {
		out = append(out, &ruleMaxFileLoc{max: *cfg.MaxFileSize})
	}
	if cfg.MaxFilesPerPackage != nil && *cfg.MaxFilesPerPackage > 0 {
		out = append(out, &ruleMaxFilesPerPackage{max: *cfg.MaxFilesPerPackage})
	}
	if isTrue(cfg.SlicePrealloc) {
		out = append(out, &ruleSlicePrealloc{})
	}
	if isTrue(cfg.ContextMissing) {
		out = append(out, &ruleContextMissing{})
	}
	if isTrue(cfg.ContextIgnored) {
		out = append(out, &ruleContextIgnored{})
	}
	return out
}
func (g *golangRuleset) All() []Rule {
	return []Rule{
		&ruleNoPkgInMethod{},
		&ruleMaxNesting{max: 4},
		&ruleMaxFileLoc{max: 1000},
		&ruleMaxFilesPerPackage{max: 50},
		&ruleSlicePrealloc{},
		// &ruleIgnoredError{}, // This rule is disabled. We should check if _ usage concernes error or not before flagging
		&ruleContextMissing{},
		&ruleContextIgnored{},
	}
}
func (g *golangRuleset) IsEnabled() bool {
	return len(g.Enabled()) > 0
}
