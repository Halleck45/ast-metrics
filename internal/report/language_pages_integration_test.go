package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ast-metrics/ast-metrics/internal/analyzer"
	"github.com/ast-metrics/ast-metrics/internal/engine"
	"github.com/ast-metrics/ast-metrics/internal/engine/csharp"
	"github.com/ast-metrics/ast-metrics/internal/engine/java"
	pb "github.com/ast-metrics/ast-metrics/pb"
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

// TestDirectoryURLSlug guards the naming of the per-folder report pages:
// analyzed paths become file-name-safe tokens (index_dir_<slug>.html).
func TestDirectoryURLSlug(t *testing.T) {
	assert.Equal(t, "internal-analyzer", directoryURLSlug("internal/analyzer"))
	assert.Equal(t, "internal-analyzer", directoryURLSlug("./internal/analyzer/"))
	assert.Equal(t, "src", directoryURLSlug("src"))
	assert.Equal(t, "my_app-v1-2", directoryURLSlug("my_app/v1.2"))
	assert.Equal(t, "a-b", directoryURLSlug("a###b"))
	assert.Equal(t, "some-folder", directoryURLSlug("some folder"))
	assert.Equal(t, "a-b", directoryURLSlug("a\\b"))
	assert.Equal(t, "root", directoryURLSlug("/"))
	assert.Equal(t, "tmp-foo-bar", directoryURLSlug("/tmp/foo/bar"))

	// no slug may leak a character that breaks a URL or a file name
	for _, dir := range []string{"a b/c#d", "../sibling", "./x/../y"} {
		slug := directoryURLSlug(dir)
		assert.NotContains(t, slug, "/")
		assert.NotContains(t, slug, "#")
		assert.NotContains(t, slug, " ")
		assert.NotContains(t, slug, ".")
		assert.NotContains(t, slug, "--")
		assert.False(t, strings.HasPrefix(slug, "-"), "slug %q must not start with a dash", slug)
		assert.False(t, strings.HasSuffix(slug, "-"), "slug %q must not end with a dash", slug)
	}
}

// TestUniqueSlug checks that two analyzed paths collapsing to the same slug
// still produce two distinct pages.
func TestUniqueSlug(t *testing.T) {
	used := map[string]bool{}
	assert.Equal(t, "src", uniqueSlug("src", used))
	assert.Equal(t, "src-2", uniqueSlug("src", used))
	assert.Equal(t, "src-3", uniqueSlug("src", used))
}

// TestBuildScopes checks the three navigation dimensions offered by the report:
// the whole project, each language, each analyzed folder.
func TestBuildScopes(t *testing.T) {
	files := []*pb.File{
		{Path: "/project/src/a.go", ProgrammingLanguage: "Golang"},
		{Path: "/project/lib/b.php", ProgrammingLanguage: "PHP"},
	}
	pa := analyzer.ProjectAggregated{
		ByProgrammingLanguage: map[string]analyzer.Aggregated{
			"Golang": {ConcernedFiles: []*pb.File{files[0]}},
			"PHP":    {ConcernedFiles: []*pb.File{files[1]}},
		},
		ByDirectory: map[string]analyzer.Aggregated{
			"/project/src": {ConcernedFiles: []*pb.File{files[0]}},
			"/project/lib": {ConcernedFiles: []*pb.File{files[1]}},
		},
	}

	scopes := buildScopes(files, pa)
	assert.Len(t, scopes, 5)

	assert.Equal(t, "all", scopes[0].Kind)
	assert.Equal(t, "", scopes[0].Suffix)
	assert.Equal(t, 2, scopes[0].FileCount)

	// languages come first, sorted, then folders, sorted
	assert.Equal(t, []string{"_Golang", "_PHP", "_dir_project-lib", "_dir_project-src"},
		[]string{scopes[1].Suffix, scopes[2].Suffix, scopes[3].Suffix, scopes[4].Suffix})
	assert.Equal(t, "directory", scopes[3].Kind)
	assert.Equal(t, "/project/lib", scopes[3].Label)
	assert.Equal(t, 1, scopes[3].FileCount)

	// each scope only keeps its own files
	assert.True(t, scopes[4].keeps(files[0]))
	assert.False(t, scopes[4].keeps(files[1]))
	assert.True(t, scopes[0].keeps(files[1]))

	// a single analyzed path produces no folder scope (ByDirectory stays empty)
	pa.ByDirectory = map[string]analyzer.Aggregated{}
	assert.Len(t, buildScopes(files, pa), 3)
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
