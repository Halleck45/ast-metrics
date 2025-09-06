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

	// Add the ruleset dynamically using the registry; avoid hardcoded defaults here
	// Verify the ruleset exists
	dummyReq := &configuration.ConfigurationRequirements{Rules: &configuration.ConfigurationRequirementsRules{}}
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

	// Ensure the category node exists in configuration.Requirements.Rules
	ensureCategory := map[string]func(){
		"architecture": func() {
			if cfg.Requirements.Rules.Architecture == nil {
				cfg.Requirements.Rules.Architecture = &configuration.ConfigurationArchitectureRules{}
			}
		},
		"volume": func() {
			if cfg.Requirements.Rules.Volume == nil {
				cfg.Requirements.Rules.Volume = &configuration.ConfigurationVolumeRules{}
			}
		},
		"complexity": func() {
			if cfg.Requirements.Rules.Complexity == nil {
				cfg.Requirements.Rules.Complexity = &configuration.ConfigurationComplexityRules{}
			}
		},
		"object-oriented-programming": func() {
			if cfg.Requirements.Rules.ObjectOrientedProgramming == nil {
				cfg.Requirements.Rules.ObjectOrientedProgramming = &configuration.ConfigurationOOPRules{}
			}
		},
	}
	if f, ok := ensureCategory[c.Name]; ok {
		f()
	}
	// Initialize empty rule keys for convenience (no numeric defaults)
	switch c.Name {
	case "volume":
		if cfg.Requirements.Rules.Volume.Loc == nil {
			cfg.Requirements.Rules.Volume.Loc = &configuration.ConfigurationDefaultRule{}
		}
		if cfg.Requirements.Rules.Volume.Lloc == nil {
			cfg.Requirements.Rules.Volume.Lloc = &configuration.ConfigurationDefaultRule{}
		}
		if cfg.Requirements.Rules.Volume.LocByMethod == nil {
			cfg.Requirements.Rules.Volume.LocByMethod = &configuration.ConfigurationDefaultRule{}
		}
		if cfg.Requirements.Rules.Volume.LlocByMethod == nil {
			cfg.Requirements.Rules.Volume.LlocByMethod = &configuration.ConfigurationDefaultRule{}
		}
	case "architecture":
		if cfg.Requirements.Rules.Architecture.Coupling == nil {
			cfg.Requirements.Rules.Architecture.Coupling = &configuration.ConfigurationCouplingRule{}
		}
		if cfg.Requirements.Rules.Architecture.AfferentCoupling == nil {
			cfg.Requirements.Rules.Architecture.AfferentCoupling = &configuration.ConfigurationDefaultRule{}
		}
		if cfg.Requirements.Rules.Architecture.EfferentCoupling == nil {
			cfg.Requirements.Rules.Architecture.EfferentCoupling = &configuration.ConfigurationDefaultRule{}
		}
		if cfg.Requirements.Rules.Architecture.Maintainability == nil {
			cfg.Requirements.Rules.Architecture.Maintainability = &configuration.ConfigurationDefaultRule{}
		}
	case "complexity":
		if cfg.Requirements.Rules.Complexity.Cyclomatic == nil {
			cfg.Requirements.Rules.Complexity.Cyclomatic = &configuration.ConfigurationDefaultRule{}
		}
	case "object-oriented-programming":
		if cfg.Requirements.Rules.ObjectOrientedProgramming.Maintainability == nil {
			cfg.Requirements.Rules.ObjectOrientedProgramming.Maintainability = &configuration.ConfigurationDefaultRule{}
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
