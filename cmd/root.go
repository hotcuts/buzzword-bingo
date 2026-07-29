package cmd

import (
	"fmt"
	"os"

	"bingo/internal/version"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "bingo",
	Short:   "Play daily bingo in your browser",
	Long:    "bingo serves a daily bingo board in your browser with persistent local state.",
	Version: version.Version,
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("bingo {{.Version}} (updated %s)\n", version.Updated))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
