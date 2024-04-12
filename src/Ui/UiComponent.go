package Ui

type UiComponent interface {
	AsTerminalElement() string
	AsHtml() string
}
