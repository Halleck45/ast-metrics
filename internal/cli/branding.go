package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// AST-style tree logo lines.
var logoLines = []string{
	`    ◉    `,
	`   ╭┴╮   `,
	`  ◉  ◎  `,
	` ╭┴╮  ╲  `,
	` ◎ ◌   ◌ `,
}

// RenderHeader returns the formatted header with logo on the left
// and project info on the right, styled like Claude CLI.
// Optional size parameters control layout: width, height.
// If too narrow or too short, the logo is hidden.
func RenderHeader(version string, colored bool, size ...int) string {
	termWidth := 0
	termHeight := 0
	if len(size) > 0 {
		termWidth = size[0]
	}
	if len(size) > 1 {
		termHeight = size[1]
	}
	compact := (termWidth > 0 && termWidth < 55) || (termHeight > 0 && termHeight < 20)
	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(cwd, home) {
		cwd = filepath.Join("~", cwd[len(home):])
	}

	slogan := "See the invisible structure of your code"

	// Info lines aligned to specific logo lines:
	// line 1: "AST Metrics"
	// line 2: slogan
	// line 4: cwd (last logo line)
	type infoLine struct {
		logoIdx int
		text    string
	}
	infoLines := []infoLine{
		{1, "AST Metrics"},
		{2, slogan},
		{4, cwd},
	}

	if !colored {
		if compact {
			return "AST Metrics\n" + slogan + "\n" + cwd
		}
		infoMap := map[int]string{}
		for _, il := range infoLines {
			infoMap[il.logoIdx] = il.text
		}
		var result []string
		for i, line := range logoLines {
			padded := fmt.Sprintf("%-12s", line)
			if text, ok := infoMap[i]; ok {
				padded += "  " + text
			}
			result = append(result, padded)
		}
		return strings.Join(result, "\n")
	}

	nodeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#73F59F")).Bold(true)
	edgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sloganStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)

	// Compact mode for narrow/short terminals: just text, no logo
	if compact {
		return titleStyle.Render("AST Metrics") + "\n" +
			sloganStyle.Render(slogan) + "\n" +
			infoStyle.Render(cwd)
	}

	styledMap := map[int]string{
		1: titleStyle.Render("AST Metrics"),
		2: sloganStyle.Render(slogan),
		4: infoStyle.Render(cwd),
	}

	pad := 12
	var result []string

	for i, line := range logoLines {
		// Colorize: nodes (◉◎◌) in green, edges (╭┴╮╲) in gray
		var coloredLine strings.Builder
		for _, ch := range line {
			s := string(ch)
			switch ch {
			case '◉', '◎', '◌':
				coloredLine.WriteString(nodeStyle.Render(s))
			case '╭', '┴', '╮', '╲':
				coloredLine.WriteString(edgeStyle.Render(s))
			default:
				coloredLine.WriteString(s)
			}
		}

		// Pad with spaces to align info column
		raw := line
		padCount := pad - len([]rune(raw))
		if padCount < 0 {
			padCount = 0
		}
		row := coloredLine.String() + strings.Repeat(" ", padCount)

		if text, ok := styledMap[i]; ok {
			row += "  " + text
		}
		result = append(result, row)
	}

	return strings.Join(result, "\n")
}

// RenderBanner returns the full banner for non-interactive mode.
func RenderBanner(version string) string {
	return RenderHeader(version, false) + "\n"
}
