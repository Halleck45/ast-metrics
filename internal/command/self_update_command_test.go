package command

import (
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSelfUpdateExecute(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("self-update not supported on Windows for this test")
	}

	// Mock http.Get
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	json := `
	{
		"url": "https://api.github.com/repos/Halleck45/ast-metrics/releases/148429686",
		"name": "v0.0.10-alpha",
		"draft": false,
		"assets": [
		  {
			"name": "ast-metrics_0.0.10-alpha_checksums.txt",
			"browser_download_url": "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_0.0.10-alpha_checksums.txt"
		  },
		  {
			"name": "ast-metrics_Darwin_arm64",
			"browser_download_url": "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_Darwin_arm64"
		  },
		  {
			"name": "ast-metrics_Darwin_x86_64",
			"browser_download_url": "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_Darwin_x86_64"
		  },
		  {
			"name": "ast-metrics_Linux_arm64",
			"browser_download_url": "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_Linux_arm64"
		  },
		  {
			"name": "ast-metrics_Linux_x86_64",
			"browser_download_url": "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_Linux_x86_64"
		  }
		]
	  }
	  `
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/Halleck45/ast-metrics/releases/latest", httpmock.NewStringResponder(200, json))
	httpmock.RegisterResponder("GET", "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_Linux_x86_64", httpmock.NewStringResponder(200, "binary"))
	httpmock.RegisterResponder("GET", "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_Darwin_arm64", httpmock.NewStringResponder(200, "binary"))
	httpmock.RegisterResponder("GET", "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_Darwin_x86_64", httpmock.NewStringResponder(200, "binary"))
	httpmock.RegisterResponder("GET", "https://github.com/Halleck45/ast-metrics/releases/download/v0.0.10-alpha/ast-metrics_Linux_arm64", httpmock.NewStringResponder(200, "binary"))

	// use custom writer to capture output
	storeStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	v := NewSelfUpdateCommand("0.0.9")
	err := v.Execute()
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	w.Close()
	out, _ := io.ReadAll(r)
	// restore the stdout
	os.Stdout = storeStdout

	// out should contains Updating to v0.0.10-alpha
	assert.Contains(t, string(out), "Updating to v0.0.10-alpha")
}

func TestSelfUpdateExecuteWhenNoCompatibleReleaseIsFound(t *testing.T) {

	// Mock http.Get
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	json := `
	{
		"url": "https://api.github.com/repos/Halleck45/ast-metrics/releases/148429686",
		"name": "v0.0.10-alpha",
		"draft": false,
		"assets": [
		 
		]
	  }
	  `
	httpmock.RegisterResponder("GET", "https://api.github.com/repos/Halleck45/ast-metrics/releases/latest", httpmock.NewStringResponder(200, json))

	// use custom writer to capture output
	storeStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	v := NewSelfUpdateCommand("0.0.9")
	err := v.Execute()
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	w.Close()
	out, _ := io.ReadAll(r)
	// restore the stdout
	os.Stdout = storeStdout

	// out should contains No update found for your platform
	assert.Contains(t, string(out), "No update found for your platform")
}
