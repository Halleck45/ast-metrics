package command

import (
	"fmt"
	"runtime"

	"github.com/pterm/pterm"
)

type VersionCommand struct {
	CurrentVersion string
}

func NewVersionCommand(currentVersion string) *VersionCommand {
	return &VersionCommand{
		CurrentVersion: currentVersion,
	}
}

func (v *VersionCommand) Execute() error {

	arch := runtime.GOARCH
	os := runtime.GOOS

	// keep always current version on first line, so it's easier to compare
	fmt.Println(v.CurrentVersion)
	fmt.Println()

	d := pterm.TableData{
		{"Current version", v.CurrentVersion},
		{"OS", os},
		{"Architecture", arch},
	}
	printer := pterm.DefaultTable.WithData(d)
	printer.Render()
	return nil
}
