package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	installScriptURL   = "https://raw.githubusercontent.com/hotcuts/buzzword-bingo/main/scripts/install.sh"
	installScriptPSURL = "https://raw.githubusercontent.com/hotcuts/buzzword-bingo/main/scripts/install.ps1"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Install the latest bingo release",
	Long: `Download and run the public install script to replace this binary
with the latest GitHub Release (macOS arm64, Linux amd64, or Windows amd64).

Same as a fresh install — inherits VERSION, INSTALL_DIR, REPO, and other
installer environment overrides when set.`,
	SilenceUsage: true,
	RunE:         runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if runtime.GOOS == "windows" {
		return runUpdateWindows()
	}
	return runUpdateUnix()
}

func runUpdateUnix() error {
	if _, err := exec.LookPath("curl"); err != nil {
		return fmt.Errorf("curl is required to update")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		return fmt.Errorf("bash is required to update")
	}

	c := exec.Command("bash", "-c", "curl -fsSL "+installScriptURL+" | bash")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	return nil
}

func runUpdateWindows() error {
	ps, err := exec.LookPath("powershell")
	if err != nil {
		ps, err = exec.LookPath("pwsh")
		if err != nil {
			return fmt.Errorf("powershell is required to update")
		}
	}

	script := fmt.Sprintf(
		"Invoke-RestMethod -Uri %q | Invoke-Expression",
		installScriptPSURL,
	)
	c := exec.Command(ps, "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	return nil
}
