package cli

import (
	"testing"

	"github.com/halleck45/ast-metrics/internal/analyzer"
	pb "github.com/halleck45/ast-metrics/pb"
	"github.com/stretchr/testify/assert"
)

func TestScreenSummaryNewScreenSummary(t *testing.T) {
	files := []*pb.File{}
	projectAggregated := analyzer.ProjectAggregated{}
	screenSummary := NewScreenSummary(true, files, projectAggregated)

	assert.Equal(t, true, screenSummary.isInteractive)
	assert.Equal(t, files, screenSummary.files)
	assert.Equal(t, projectAggregated, screenSummary.projectAggregated)
}

func TestScreenSummaryGetScreenName(t *testing.T) {
	screenSummary := ScreenSummary{}
	assert.Equal(t, "Overview", screenSummary.GetScreenName())
}

func TestScreenSummaryGetModel(t *testing.T) {
	files := []*pb.File{}
	projectAggregated := analyzer.ProjectAggregated{}
	screenSummary := NewScreenSummary(true, files, projectAggregated)
	model := screenSummary.GetModel()

	modelScreenSummary, ok := model.(modelScreenSummary)
	assert.True(t, ok)
	assert.Equal(t, files, modelScreenSummary.files)
	assert.Equal(t, projectAggregated, modelScreenSummary.projectAggregated)
}

func TestScreenSummaryModelScreenSummaryInit(t *testing.T) {
	model := modelScreenSummary{}
	assert.Nil(t, model.Init())
}

func TestScreenSummaryModelScreenSummaryUpdate(t *testing.T) {
	model := modelScreenSummary{}
	newModel, cmd := model.Update("q")

	newScreenHome, ok := newModel.(modelScreenSummary)
	assert.True(t, ok)
	assert.Nil(t, cmd)
	assert.Equal(t, model.files, newScreenHome.files)
	assert.Equal(t, model.projectAggregated, newScreenHome.projectAggregated)
}

func TestScreenSummaryModelScreenSummaryView(t *testing.T) {
	model := modelScreenSummary{
		files: []*pb.File{},
		projectAggregated: analyzer.ProjectAggregated{
			ByClass:  analyzer.Aggregated{},
			Combined: analyzer.Aggregated{},
		},
	}

	view := model.View()
	// Assert contains cards
	assert.Contains(t, view, "🔴 < 64")
	assert.Contains(t, view, "Average LOC per method")
}
