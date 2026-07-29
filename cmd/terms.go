package cmd

import (
	"fmt"

	"bingo/internal/config"
	"bingo/internal/terms"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(termsCmd)
	termsCmd.AddCommand(termsSetCmd)
	termsCmd.AddCommand(termsAddCmd)
	termsCmd.AddCommand(termsRemoveCmd)
	termsCmd.AddCommand(termsResetCmd)
}

var termsCmd = &cobra.Command{
	Use:   "terms",
	Short: "Manage bingo terms (defaults or a custom file)",
}

var termsSetCmd = &cobra.Command{
	Use:          "set <file>",
	Short:        "Set or replace the custom terms file",
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		n, err := terms.SetFile(cfg, args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Using %d terms from %s → %s\n", n, args[0], cfg.TermsPath)
		return nil
	},
}

var termsAddCmd = &cobra.Command{
	Use:          "add <term>...",
	Short:        "Add term(s) to the custom list (seeds defaults if needed)",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		n, err := terms.Add(cfg, args...)
		if err != nil {
			return err
		}
		fmt.Printf("Added. %d terms in %s\n", n, cfg.TermsPath)
		return nil
	},
}

var termsRemoveCmd = &cobra.Command{
	Use:          "remove <term>...",
	Short:        "Remove term(s) from the custom list",
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		n, err := terms.Remove(cfg, args...)
		if err != nil {
			return err
		}
		fmt.Printf("Removed. %d terms in %s\n", n, cfg.TermsPath)
		return nil
	},
}

var termsResetCmd = &cobra.Command{
	Use:          "reset",
	Short:        "Remove custom terms and use embedded defaults",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		if err := terms.Reset(cfg); err != nil {
			return err
		}
		fmt.Println("Custom terms removed. Play will use embedded defaults.")
		return nil
	},
}