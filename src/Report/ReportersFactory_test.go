package Report

import (
	"testing"

	"github.com/halleck45/ast-metrics/src/Configuration"
	"github.com/stretchr/testify/assert"
)

func TestShouldAbleToFactoryReporters(t *testing.T) {
	t.Run("Should able to factory enabled reports", func(t *testing.T) {
		configuration := Configuration.Configuration{}
		configuration.Reports = Configuration.ConfigurationReport{
			Html:        "foo",
			Markdown:    "foo.md",
			Json:        "foo.json",
			OpenMetrics: "foo.txt",
		}

		factory := ReportersFactory{Configuration: &configuration}
		reporters := factory.Factory(&configuration)
		assert.Equal(t, 4, len(reporters))
	})

	t.Run("Should able to factory when no report is configured", func(t *testing.T) {
		configuration := Configuration.Configuration{}
		factory := ReportersFactory{Configuration: &configuration}
		reporters := factory.Factory(&configuration)
		assert.Equal(t, 0, len(reporters))
	})

	t.Run("Should able to factory when only html report is configured", func(t *testing.T) {
		configuration := Configuration.Configuration{}
		configuration.Reports = Configuration.ConfigurationReport{
			Html:     "foo",
			Markdown: "",
		}

		factory := ReportersFactory{Configuration: &configuration}
		reporters := factory.Factory(&configuration)
		assert.Equal(t, 1, len(reporters))
		assert.Equal(t, "foo", reporters[0].(*HtmlReportGenerator).ReportPath)
	})

	t.Run("Should able to factory when only markdown report is configured", func(t *testing.T) {
		configuration := Configuration.Configuration{}
		configuration.Reports = Configuration.ConfigurationReport{
			Html:     "",
			Markdown: "foo.md",
		}

		factory := ReportersFactory{Configuration: &configuration}
		reporters := factory.Factory(&configuration)
		assert.Equal(t, 1, len(reporters))
		assert.Equal(t, "foo.md", reporters[0].(*MarkdownReportGenerator).ReportPath)
	})
}
