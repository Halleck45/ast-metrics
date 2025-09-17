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
	Sarif       string `yaml:"sarif,omitempty"`
}

// function HasReports() bool {
func (c *ConfigurationReport) HasReports() bool {
	return c.Html != "" || c.Markdown != "" || c.Json != "" || c.OpenMetrics != "" || c.Sarif != ""
}

type ConfigurationRequirements struct {
	Rules   *ConfigurationRequirementsRules `yaml:"rules"`
	Exclude []string                        `yaml:"exclude,omitempty"`
}

type ConfigurationCouplingRule struct {
	Forbidden []struct {
		From string `yaml:"from"`
		To   string `yaml:"to"`
	} `yaml:"forbidden,omitempty"`
}

type ConfigurationArchitectureRules struct {
	Coupling              *ConfigurationCouplingRule `yaml:"coupling,omitempty"`
	AfferentCoupling      *int                       `yaml:"max_afferent_coupling,omitempty"`
	EfferentCoupling      *int                       `yaml:"max_efferent_coupling,omitempty"`
	Maintainability       *int                       `yaml:"min_maintainability,omitempty"`
	NoCircularDependencies *bool                      `yaml:"no_circular_dependencies,omitempty"`
	MaxResponsibilities   *int                       `yaml:"max_responsibilities,omitempty"`
	NoGodClass            *bool                      `yaml:"no_god_class,omitempty"`
}

type ConfigurationVolumeRules struct {
	Loc                  *int `yaml:"max_loc,omitempty"`
	Lloc                 *int `yaml:"max_logical_loc,omitempty"`
	LocByMethod          *int `yaml:"max_loc_by_method,omitempty"`
	LlocByMethod         *int `yaml:"max_logical_loc_by_method,omitempty"`
	MaxMethodsPerClass   *int `yaml:"max_methods_per_class,omitempty"`
	MaxSwitchCases       *int `yaml:"max_switch_cases,omitempty"`
	MaxParametersPerMethod *int `yaml:"max_parameters_per_method,omitempty"`
	MaxNestedBlocks      *int `yaml:"max_nested_blocks,omitempty"`
	MaxPublicMethods     *int `yaml:"max_public_methods,omitempty"`
}

type ConfigurationComplexityRules struct {
	Cyclomatic *int `yaml:"max_cyclomatic,omitempty"`
}

type ConfigurationOOPRules struct {
	Maintainability *int `yaml:"min_maintainability,omitempty"`
}

type ConfigurationRequirementsRules struct {
	// New nested rulesets
	Architecture              *ConfigurationArchitectureRules `yaml:"architecture,omitempty"`
	Volume                    *ConfigurationVolumeRules       `yaml:"volume,omitempty"`
	Complexity                *ConfigurationComplexityRules   `yaml:"complexity,omitempty"`
	ObjectOrientedProgramming *ConfigurationOOPRules          `yaml:"object-oriented-programming,omitempty"`
	Golang                    *ConfigurationGolangRuleset     `yaml:"golang,omitempty"`

	// Legacy flat rules support for backward compatibility
	CyclomaticLegacy *ConfigurationDefaultRule `yaml:"cyclomatic_complexity,omitempty"`
}

// ConfigurationGolangRuleset toggles for Golang-specific best-practice rules (per-rule)
// If a field is set to true, the corresponding rule is enabled. Omitting or false disables it.
type ConfigurationGolangRuleset struct {
	NoPackageNameInMethod *bool `yaml:"no_package_name_in_method,omitempty"`
	MaxNesting            *int  `yaml:"max_nesting,omitempty"`
	MaxFileSize           *int  `yaml:"max_file_size,omitempty"`
	MaxFilesPerPackage    *int  `yaml:"max_files_per_package,omitempty"`
	SlicePrealloc         *bool `yaml:"slice_prealloc,omitempty"`
	ContextMissing        *bool `yaml:"context_missing,omitempty"`
	ContextIgnored        *bool `yaml:"context_ignored,omitempty"`
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
