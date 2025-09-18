package report

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
)

// OpenHtmlReport opens the HTML report in the default browser
func OpenHtmlReport(reportPath string) error {
	// Ensure the report path exists
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		return fmt.Errorf("HTML report not found at: %s", reportPath)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(reportPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Convert to file:// URL
	fileURL := "file://" + strings.ReplaceAll(absPath, "\\", "/")

	// Determine the command to open the browser based on the OS
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", fileURL)
	case "darwin":
		cmd = exec.Command("open", fileURL)
	case "linux":
		// Try xdg-open first, then fallback to other methods
		if _, err := exec.LookPath("xdg-open"); err == nil {
			cmd = exec.Command("xdg-open", fileURL)
		} else if _, err := exec.LookPath("firefox"); err == nil {
			cmd = exec.Command("firefox", fileURL)
		} else if _, err := exec.LookPath("google-chrome"); err == nil {
			cmd = exec.Command("google-chrome", fileURL)
		} else if _, err := exec.LookPath("chromium-browser"); err == nil {
			cmd = exec.Command("chromium-browser", fileURL)
		} else {
			return fmt.Errorf("no suitable browser found. Please open manually: %s", absPath)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Execute the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser: %v. Please open manually: %s", err, absPath)
	}

	// Show success message
	pterm.Success.Printf("🌐 HTML report opened in browser: %s\n", absPath)
	return nil
}

// GetOpenInstructions returns platform-specific instructions for opening the HTML report
func GetOpenInstructions(reportPath string) string {
	absPath, err := filepath.Abs(reportPath)
	if err != nil {
		absPath = reportPath
	}

	fileURL := "file://" + strings.ReplaceAll(absPath, "\\", "/")

	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("💡 To open the report: start %s", fileURL)
	case "darwin":
		return fmt.Sprintf("💡 To open the report: open %s", fileURL)
	case "linux":
		return fmt.Sprintf("💡 To open the report: xdg-open %s", fileURL)
	default:
		return fmt.Sprintf("💡 To open the report: %s", absPath)
	}
}
