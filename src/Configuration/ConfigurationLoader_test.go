package Configuration

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadsFileExists(t *testing.T) {
	// Create a temporary file
	tempFile, _ := ioutil.TempFile("", "test")
	defer os.Remove(tempFile.Name())

	// Write some YAML to the file
	yamlData := `
sources: 
  - ./myfolder1
exclude: []

reports:
  html: ./build/report
  markdown: ./build/report.md

requirements:
  rules:
    cyclomatic_complexity:
      max: 10
      excludes: []

    lines_of_code:
      max: 100
      excludes: []

`
	os.WriteFile(tempFile.Name(), []byte(yamlData), 0644)

	// Create a ConfigurationLoader with the temp file name
	loader := &ConfigurationLoader{FilenameToChecks: []string{tempFile.Name()}}

	// Call Loads
	cfg, err := loader.Loads(&Configuration{})

	// Assert no error
	assert.NoError(t, err)

	// Assert the configuration was loaded correctly
	assert.Equal(t, []string{"./myfolder1"}, cfg.SourcesToAnalyzePath)
}

func TestLoadsFileDoesNotExist(t *testing.T) {
	// Create a ConfigurationLoader with a non-existent file name
	loader := &ConfigurationLoader{FilenameToChecks: []string{"non_existent_file.yaml"}}

	// Call Loads
	cfg, err := loader.Loads(&Configuration{})

	// Assert no error
	assert.NoError(t, err)

	// Assert the configuration is empty
	assert.Equal(t, &Configuration{}, cfg)
}

func TestLoadsErrorDecodingYAML(t *testing.T) {
	// Create a temporary file
	tempFile, _ := ioutil.TempFile("", "test")
	defer os.Remove(tempFile.Name())

	// Write some invalid YAML to the file
	invalidYAML := `key: value:`
	ioutil.WriteFile(tempFile.Name(), []byte(invalidYAML), 0644)

	// Create a ConfigurationLoader with the temp file name
	loader := &ConfigurationLoader{FilenameToChecks: []string{tempFile.Name()}}

	// Call Loads
	cfg, err := loader.Loads(&Configuration{})

	// Assert there was an error
	assert.Error(t, err)

	// Assert the configuration is empty
	assert.Equal(t, &Configuration{}, cfg)
}

func TestCreateDefaultFile(t *testing.T) {
	loader := NewConfigurationLoader()

	err := loader.CreateDefaultFile()
	assert.NoError(t, err)

	// Check if the file was created
	_, err = os.Stat(".ast-metrics.yaml")
	assert.NoError(t, err)

	// Clean up
	os.Remove(".ast-metrics.yaml")
}
