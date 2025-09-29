package ui

type UiComponent interface {
	AsTerminalElement() string
	AsHtml() string
}
