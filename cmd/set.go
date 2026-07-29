package cmd

import (
	"fmt"
	"strings"

	"bingo/internal/config"
	"bingo/internal/profile"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setNameCmd)
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set player settings",
}

var setNameCmd = &cobra.Command{
	Use:          "name <name>",
	Short:        "Set your player name (shown on the board)",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		name := strings.Join(args, " ")
		if err := profile.Set(cfg, name); err != nil {
			return err
		}
		fmt.Printf("Name set to %q\n", strings.TrimSpace(name))
		return nil
	},
}