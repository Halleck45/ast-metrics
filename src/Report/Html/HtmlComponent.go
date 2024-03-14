package Report

type Component interface {

	// Returns the content of the component
	RenderHtml() string
}
