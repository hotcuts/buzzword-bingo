package cmd

import (
	"fmt"

	"bingo/internal/version"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and last update date",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bingo %s (updated %s)\n", version.Version, version.Updated)
	},
}
