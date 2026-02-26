package cli

import (
	"github.com/pterm/pterm"
)

// MoonSpinnerSequence is the moon-phase animation frames.
var MoonSpinnerSequence = []string{
	"🌑·····",
	"·🌒····",
	"··🌓···",
	"···🌔··",
	"····🌕·",
	"···🌖··",
	"··🌗···",
	"·🌘····",
}

// NewMoonSpinner creates a pterm spinner with the moon-phase animation.
func NewMoonSpinner(message string) (*pterm.SpinnerPrinter, error) {
	return pterm.DefaultSpinner.
		WithSequence(MoonSpinnerSequence...).
		WithRemoveWhenDone(true).
		Start(message)
}
