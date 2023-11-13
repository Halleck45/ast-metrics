package Php

import (
	"os"
	"testing"

	"github.com/halleck45/ast-metrics/src/Configuration"
)

func TestPhpRunner(t *testing.T) {
	t.Run("CheckIfRequired Without PHP files", func(t *testing.T) {

		// create a temporary directory
		tempDir := t.TempDir()
		configuration := Configuration.NewConfiguration()
		configuration.SourcesToAnalyzePath = []string{tempDir}

		// put a javascript file in it
		if err := os.WriteFile(tempDir+"/test.js", []byte("console.log('hello world');"), 0644); err != nil {
			t.Fatal(err)
		}

		// create a PhpRunner
		runner := PhpRunner{configuration: configuration}

		// check if required
		if runner.IsRequired() {
			t.Fatal("PHP runner should not be required")
		}
	})

	t.Run("CheckIfRequired with PHP files", func(t *testing.T) {

		// create a temporary directory
		tempDir := t.TempDir()
		configuration := Configuration.NewConfiguration()
		configuration.SourcesToAnalyzePath = []string{tempDir}

		// put a javascript file in it
		if err := os.WriteFile(tempDir+"/test.php", []byte("<? echo 1;"), 0644); err != nil {
			t.Fatal(err)
		}

		// create a PhpRunner
		runner := PhpRunner{configuration: configuration}

		// check if required
		if !runner.IsRequired() {
			t.Fatal("PHP runner should be required")
		}
	})
}
