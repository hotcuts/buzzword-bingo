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
	setCmd.AddCommand(setPeriodCmd)
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

var setPeriodCmd = &cobra.Command{
	Use:   "period [daily|weekly]",
	Short: "Set or show how often the board resets (daily or ISO weekly)",
	Long: `Set the board reset period to "daily" or "weekly" (ISO calendar week, Monday start).
With no argument, prints the current period (defaults to daily).`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			period, err := profile.GetPeriod(cfg)
			if err != nil {
				return err
			}
			fmt.Printf("Board resets %s\n", period)
			return nil
		}
		period, err := profile.ParsePeriod(args[0])
		if err != nil {
			return err
		}
		if err := profile.SetPeriod(cfg, period); err != nil {
			return err
		}
		fmt.Printf("Board reset period set to %q\n", period)
		return nil
	},
}