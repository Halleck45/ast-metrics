package Configuration

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigurationLoader struct {
	FilenameToChecks []string
}

func NewConfigurationLoader() *ConfigurationLoader {
	return &ConfigurationLoader{
		FilenameToChecks: []string{
			".ast-metrics.yaml",
			".ast-metrics.dist.yaml",
		},
	}
}

func (c *ConfigurationLoader) Loads(cfg *Configuration) (*Configuration, error) {
	// Load configuration file
	for _, filename := range c.FilenameToChecks {

		if _, err := os.Stat(filename); err == nil {

			// Load configuration
			f, err := os.Open(filename)
			if err != nil {
				return cfg, err
			}
			defer f.Close()

			decoder := yaml.NewDecoder(f)
			err = decoder.Decode(&cfg)
			if err != nil {
				return cfg, err
			}

			return cfg, nil
		}
	}

	return cfg, nil
}

func (c *ConfigurationLoader) CreateDefaultFile() error {
	if len(c.FilenameToChecks) == 0 {
		return errors.New("No filename to check")
	}
	filename := c.FilenameToChecks[0]

	// Create default configuration file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = f.WriteString(`# AST Metrics configuration file
# This file is used to configure AST Metrics
# You can find more information at https://github.com/Halleck45/ast-metrics/

# Sources to analyze. You can add multiple sources
sources:
  - ./

# Exclude patterns (list of regular expressions. When a file matches one of these patterns, it is not analyzed)
exclude:
  - /vendor/
  - /node_modules/

# Reports to generate
reports:
  html: ./build/report
  markdown: ./build/report.md

# Requirements. If a file does not meet these requirements, it will be reported
requirements:
  rules:

	# Complexity of the code
    cyclomatic_complexity:
	  max: 10
	  exclude: []

	# Number of lines of code
	loc:
	  max: 100
	  exclude: []
`)

	if err != nil {
		return err
	}

	return nil
}
