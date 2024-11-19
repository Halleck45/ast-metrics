package Report

import (
	"github.com/halleck45/ast-metrics/src/Configuration"
)

type ReportersFactory struct {
	Configuration *Configuration.Configuration
}

func (v *ReportersFactory) Factory(configuration *Configuration.Configuration) []Reporter {
	reporters := []Reporter{}
	if v.Configuration.Reports.HasReports() {
		if v.Configuration.Reports.Html != "" {
			reporters = append(reporters, NewHtmlReportGenerator(v.Configuration.Reports.Html))
		}
		if v.Configuration.Reports.Markdown != "" {
			reporters = append(reporters, NewMarkdownReportGenerator(v.Configuration.Reports.Markdown))
		}
		if v.Configuration.Reports.Json != "" {
			reporters = append(reporters, NewJsonReportGenerator(v.Configuration.Reports.Json))
		}
		if v.Configuration.Reports.OpenMetrics != "" {
			reporters = append(reporters, NewOpenMetricsReportGenerator(v.Configuration.Reports.OpenMetrics))
		}
	}

	return reporters
}
