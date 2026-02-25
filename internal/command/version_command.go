package command

import (
	"fmt"
	"runtime"

	"github.com/halleck45/ast-metrics/internal/cli"
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
	fmt.Print(cli.ScreenHeader("Version"))
	fmt.Println()

	fmt.Printf("  %-20s %s\n", "Current version", v.CurrentVersion)
	fmt.Printf("  %-20s %s\n", "OS", runtime.GOOS)
	fmt.Printf("  %-20s %s\n", "Architecture", runtime.GOARCH)
	return nil
}
