package configuration

import (
	"os"
	"path/filepath"

	storage "github.com/halleck45/ast-metrics/internal/storage"
)

type Configuration struct {
	// The path to the sources to analyze
	SourcesToAnalyzePath []string `yaml:"sources"`

	// Exclude patterns (list of regular expressions. When a file matches one of these patterns, it is not analyzed)
	ExcludePatterns []string `yaml:"exclude"`

	// Reports
	Reports ConfigurationReport `yaml:"reports,omitempty"`

	// Requirements
	Requirements *ConfigurationRequirements `yaml:"requirements,omitempty"`

	Watching bool `yaml:"watching,omitempty"`

	// if not empty, compare the current analysis with the one in this branch / commit
	CompareWith string `yaml:"comparewith,omitempty"`

	// Location of cache files
	Storage *storage.Workdir `yaml:"-"`

	IsComingFromConfigFile bool `yaml:"-"`
}

type ConfigurationReport struct {
	Html        string `yaml:"html,omitempty"`
	Markdown    string `yaml:"markdown,omitempty"`
	Json        string `yaml:"json,omitempty"`
	OpenMetrics string `yaml:"openmetrics,omitempty"`
}

// function HasReports() bool {
func (c *ConfigurationReport) HasReports() bool {
	return c.Html != "" || c.Markdown != "" || c.Json != "" || c.OpenMetrics != ""
}

type ConfigurationRequirements struct {
	Rules *ConfigurationRequirementsRules `yaml:"rules"`

	FailOnError bool `yaml:"fail_on_error"`
}

type ConfigurationCouplingRule struct {
	Forbidden []struct {
		From string `yaml:"from"`
		To   string `yaml:"to"`
	} `yaml:"forbidden,omitempty"`
}

type ConfigurationArchitectureRules struct {
	Coupling        *ConfigurationCouplingRule `yaml:"coupling,omitempty"`
	AfferentCoupling *ConfigurationDefaultRule `yaml:"afferent_coupling,omitempty"`
	EfferentCoupling *ConfigurationDefaultRule `yaml:"efferent_coupling,omitempty"`
	Maintainability  *ConfigurationDefaultRule `yaml:"maintainability,omitempty"`
}

type ConfigurationVolumeRules struct {
	Loc           *ConfigurationDefaultRule `yaml:"loc,omitempty"`
	Lloc          *ConfigurationDefaultRule `yaml:"lloc,omitempty"`
	LocByMethod   *ConfigurationDefaultRule `yaml:"loc_by_method,omitempty"`
	LlocByMethod  *ConfigurationDefaultRule `yaml:"lloc_by_method,omitempty"`
}

type ConfigurationComplexityRules struct {
	Cyclomatic *ConfigurationDefaultRule `yaml:"cyclomatic_complexity,omitempty"`
}

type ConfigurationOOPRules struct {
	Maintainability *ConfigurationDefaultRule `yaml:"maintainability,omitempty"`
}

type ConfigurationRequirementsRules struct {
	// New nested rulesets
	Architecture              *ConfigurationArchitectureRules `yaml:"architecture,omitempty"`
	Volume                    *ConfigurationVolumeRules       `yaml:"volume,omitempty"`
	Complexity                *ConfigurationComplexityRules   `yaml:"complexity,omitempty"`
	ObjectOrientedProgramming *ConfigurationOOPRules          `yaml:"object-oriented-programming,omitempty"`

	// Legacy flat rules (backward compatibility)
	CyclomaticComplexity *ConfigurationDefaultRule  `yaml:"cyclomatic_complexity,omitempty"`
	Loc                  *ConfigurationDefaultRule  `yaml:"loc,omitempty"`
	Maintainability      *ConfigurationDefaultRule  `yaml:"maintainability,omitempty"`
	Coupling             *ConfigurationCouplingRule `yaml:"coupling,omitempty"`
}

type ConfigurationDefaultRule struct {
	Max             int      `yaml:"max"`
	Min             int      `yaml:"min"`
	ExcludePatterns []string `yaml:"exclude"`
}

func NewConfiguration() *Configuration {
	return &Configuration{
		SourcesToAnalyzePath:   []string{},
		ExcludePatterns:        []string{"/vendor/", "/node_modules/", "/.git/", "/.idea/", "/tests/", "/Tests/", "/test/", "/Test/", "/spec/", "/Spec/", "/_ide_helper/"},
		Watching:               false,
		CompareWith:            "",
		Storage:                storage.Default(),
		IsComingFromConfigFile: false,
		Requirements:           &ConfigurationRequirements{FailOnError: true},
	}
}

func NewConfigurationRequirements() *ConfigurationRequirements {
	return &ConfigurationRequirements{
		Rules: &ConfigurationRequirementsRules{
			Architecture:              &ConfigurationArchitectureRules{},
			Volume:                    &ConfigurationVolumeRules{},
			Complexity:                &ConfigurationComplexityRules{},
			ObjectOrientedProgramming: &ConfigurationOOPRules{},
		},
	}
}

func (c *Configuration) SetSourcesToAnalyzePath(paths []string) error {

	(*c).SourcesToAnalyzePath = []string{}

	// foreach path, make it absolute
	for i := range paths {

		// ensure path exists
		if _, err := os.Stat(paths[i]); err != nil {
			return err
		}

		// make path absolute
		if !filepath.IsAbs(paths[i]) {
			var err error
			paths[i], err = filepath.Abs(paths[i])
			if err != nil {
				return err
			}
		}

		// ensure path exists
		if _, err := os.Stat(paths[i]); err != nil {
			return err
		}

		// Remove trailing slash
		paths[i] = filepath.Clean(paths[i])
	}

	(*c).SourcesToAnalyzePath = paths

	return nil
}

func (c *Configuration) SetExcludePatterns(patterns []string) {
	// Ensure patterns are valid regular expressions
	// @todo
	(*c).ExcludePatterns = patterns
}
