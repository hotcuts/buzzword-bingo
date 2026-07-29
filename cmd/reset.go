package cmd

import (
	"fmt"

	"bingo/internal/config"
	"bingo/internal/session"
	"bingo/internal/terms"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(resetCmd)
	resetCmd.AddCommand(resetBoardCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Restore bingo config to defaults (name, terms, session, wins)",
	Long: `Remove local bingo settings under ~/.config/bingo so the app
starts fresh: embedded terms, no player name, empty wins, new session on play.

Use "bingo reset board" to only reshuffle today's board.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		if err := config.ResetAll(cfg); err != nil {
			return err
		}
		fmt.Printf("Config reset to defaults (%s).\n", cfg.Dir)
		return nil
	},
}

var resetBoardCmd = &cobra.Command{
	Use:          "board",
	Short:        "Reshuffle today's board (keeps wins and settings)",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Ensure()
		if err != nil {
			return err
		}
		pool, err := terms.Load(cfg)
		if err != nil {
			return err
		}
		store := session.NewStore(cfg)
		game, err := store.LoadOrCreate(pool)
		if err != nil {
			return err
		}
		if _, err := store.ResetTally(game, pool); err != nil {
			return err
		}
		fmt.Println("Board reshuffled. Wins and settings unchanged.")
		return nil
	},
}
