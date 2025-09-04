package command

import (
	"errors"
	"fmt"
	"os"

	"github.com/halleck45/ast-metrics/internal/analyzer/ruleset"
	"github.com/halleck45/ast-metrics/internal/configuration"
	"gopkg.in/yaml.v3"
)

type RulesetAddCommand struct{ Name string }

func NewRulesetAddCommand(name string) *RulesetAddCommand { return &RulesetAddCommand{Name: name} }

func (c *RulesetAddCommand) Execute() error {
	if c.Name == "" {
		return errors.New("ruleset name is required")
	}

	// Load config file (prefer .ast-metrics.yaml then .ast-metrics.dist.yaml)
	loader := configuration.NewConfigurationLoader()
	cfg := configuration.NewConfiguration()
	cfg, _ = loader.Loads(cfg)
	if !cfg.IsComingFromConfigFile {
		// Create default config file if none exists
		if err := loader.CreateDefaultFile(); err != nil {
			return err
		}
		// Reload
		cfg, _ = loader.Loads(cfg)
	}

	// Ensure requirements structure exists
	if cfg.Requirements == nil {
		cfg.Requirements = &configuration.ConfigurationRequirements{Rules: &configuration.ConfigurationRequirementsRules{}}
	}
	if cfg.Requirements.Rules == nil {
		cfg.Requirements.Rules = &configuration.ConfigurationRequirementsRules{}
	}

	// Add the ruleset and initialize functional defaults only if missing
	switch c.Name {
	case "architecture":
		if cfg.Requirements.Rules.Architecture == nil {
			cfg.Requirements.Rules.Architecture = &configuration.ConfigurationArchitectureRules{}
		}
		if cfg.Requirements.Rules.Architecture.Coupling == nil {
			cfg.Requirements.Rules.Architecture.Coupling = &configuration.ConfigurationCouplingRule{}
			cfg.Requirements.Rules.Architecture.Coupling.Forbidden = append(cfg.Requirements.Rules.Architecture.Coupling.Forbidden, struct {
				From string `yaml:"from"`
				To   string `yaml:"to"`
			}{From: "Model", To: "Controller"})
		}
	case "volume":
		if cfg.Requirements.Rules.Volume == nil {
			cfg.Requirements.Rules.Volume = &configuration.ConfigurationVolumeRules{}
		}
		if cfg.Requirements.Rules.Volume.Loc == nil {
			cfg.Requirements.Rules.Volume.Loc = &configuration.ConfigurationDefaultRule{Max: 100, ExcludePatterns: []string{}}
		}
	case "complexity":
		if cfg.Requirements.Rules.Complexity == nil {
			cfg.Requirements.Rules.Complexity = &configuration.ConfigurationComplexityRules{}
		}
		if cfg.Requirements.Rules.Complexity.Cyclomatic == nil {
			cfg.Requirements.Rules.Complexity.Cyclomatic = &configuration.ConfigurationDefaultRule{Max: 10, ExcludePatterns: []string{}}
		}
	case "object-oriented-programming":
		if cfg.Requirements.Rules.ObjectOrientedProgramming == nil {
			cfg.Requirements.Rules.ObjectOrientedProgramming = &configuration.ConfigurationOOPRules{}
		}
		if cfg.Requirements.Rules.ObjectOrientedProgramming.Maintainability == nil {
			cfg.Requirements.Rules.ObjectOrientedProgramming.Maintainability = &configuration.ConfigurationDefaultRule{Min: 85}
		}
	default:
		// Check against existing registry names
		dummyReq := &configuration.ConfigurationRequirements{Rules: &configuration.ConfigurationRequirementsRules{Architecture: &configuration.ConfigurationArchitectureRules{}, Volume: &configuration.ConfigurationVolumeRules{}}}
		found := false
		for _, s := range ruleset.Registry(dummyReq).AllRulesets() {
			if s.Category() == c.Name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown ruleset '%s'", c.Name)
		}
	}

	// Save back to file
	// Check if any loader.FilenameToChecks file exists
	filename := loader.FilenameToChecks[0]
	for _, fn := range loader.FilenameToChecks {
		if _, err := os.Stat(fn); err == nil {
			filename = fn
			break
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := yaml.NewEncoder(file)
	enc.SetIndent(2)
	if err := enc.Encode(cfg); err != nil {
		return err
	}
	fmt.Printf("Ruleset '%s' added in %s\n", c.Name, filename)
	return nil
}
