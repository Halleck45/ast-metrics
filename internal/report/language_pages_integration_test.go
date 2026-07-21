package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	"github.com/halleck45/ast-metrics/internal/engine"
	"github.com/halleck45/ast-metrics/internal/engine/csharp"
	"github.com/halleck45/ast-metrics/internal/engine/java"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/stretchr/testify/assert"
)

// TestLanguageURLSlug guards the URL-safe naming of per-language report files:
// "C#" must not leak a "#" into file names or hrefs (the browser would treat
// it as a fragment delimiter).
func TestLanguageURLSlug(t *testing.T) {
	assert.Equal(t, "CSharp", languageURLSlug("C#"))
	assert.Equal(t, "Java", languageURLSlug("Java"))
	assert.Equal(t, "Golang", languageURLSlug("Golang"))
	assert.Equal(t, "PHP", languageURLSlug("PHP"))
}

// TestHtmlReportFullPipelineJavaCSharp runs the whole pipeline (parse with the
// Java and C# engines, analyze, aggregate, generate the HTML report) and
// verifies the report files for both languages.
func TestHtmlReportFullPipelineJavaCSharp(t *testing.T) {
	javaSrc := `
package com.example;

import java.util.List;

public class Calculator {
    public int add(int a, int b) {
        return a + b;
    }

    public String classify(int x) {
        if (x > 0) {
            return "pos";
        } else if (x == 0) {
            return "zero";
        } else {
            return "neg";
        }
    }
}
`
	csSrc := `
using System;

namespace App.Services
{
    public class UserService
    {
        public string Find(string[] users, string prefix)
        {
            foreach (var u in users)
            {
                if (u.StartsWith(prefix))
                {
                    return u;
                }
            }
            return null;
        }
    }
}
`
	javaFile, err := engine.CreateTestFileWithCode(&java.JavaRunner{}, javaSrc)
	assert.Nil(t, err)
	csFile, err := engine.CreateTestFileWithCode(&csharp.CSharpRunner{}, csSrc)
	assert.Nil(t, err)

	files := []*pb.File{javaFile, csFile}
	analyzer.AnalyzeFiles(files, nil)
	pa := analyzer.NewAggregator(files, nil).Aggregates()

	// both languages must be aggregated under their display names
	assert.Contains(t, pa.ByProgrammingLanguage, "Java")
	assert.Contains(t, pa.ByProgrammingLanguage, "C#")

	reportDir, err := os.MkdirTemp("", "ast-metrics-report-*")
	assert.Nil(t, err)
	defer os.RemoveAll(reportDir)

	generator := NewHtmlReportGenerator(reportDir)
	_, err = generator.Generate(files, pa)
	assert.Nil(t, err)

	// per-language pages exist, with URL-safe names
	for _, page := range []string{"index_Java.html", "index_CSharp.html", "explorer_Java.html", "explorer_CSharp.html"} {
		_, statErr := os.Stat(filepath.Join(reportDir, page))
		assert.Nil(t, statErr, "Expected page %s to exist", page)
	}
	_, statErr := os.Stat(filepath.Join(reportDir, "data", "data_CSharp.js"))
	assert.Nil(t, statErr, "Expected data/data_CSharp.js to exist")

	// no generated file may contain "#" in its name
	walkErr := filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		assert.NotContains(t, filepath.Base(path), "#", "Report file names must be URL-safe")
		return nil
	})
	assert.Nil(t, walkErr)

	// language tab links must use the slug, not the raw language name
	indexContent, err := os.ReadFile(filepath.Join(reportDir, "index.html"))
	assert.Nil(t, err)
	assert.Contains(t, string(indexContent), "index_CSharp.html", "Language tab must link to the slugged page")
	assert.NotContains(t, string(indexContent), "index_C#.html", "No raw # in hrefs")

	// qualified names use native separators in the report data
	dataJava, err := os.ReadFile(filepath.Join(reportDir, "data", "data_Java.js"))
	assert.Nil(t, err)
	assert.Contains(t, string(dataJava), "com.example.Calculator", "Java qualified names use dots")

	dataCs, err := os.ReadFile(filepath.Join(reportDir, "data", "data_CSharp.js"))
	assert.Nil(t, err)
	assert.Contains(t, string(dataCs), "App.Services.UserService", "C# qualified names use dots")

	// the C# page still displays the language with its real name
	csPage, err := os.ReadFile(filepath.Join(reportDir, "index_CSharp.html"))
	assert.Nil(t, err)
	assert.True(t, strings.Contains(string(csPage), "C#"), "Display name stays C#")
}
