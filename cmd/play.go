package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"bingo/internal/config"
	"bingo/internal/server"
	"bingo/internal/session"
	"bingo/internal/terms"
	"bingo/internal/updatecheck"
	"bingo/internal/version"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(playCmd)
}

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Start today's bingo game in the browser",
	RunE:  runPlay,
}

func runPlay(cmd *cobra.Command, args []string) error {
	cfg, err := config.Ensure()
	if err != nil {
		return err
	}
	go updatecheck.MaybeNotify(version.Version)

	pool, err := terms.Load(cfg)
	if err != nil {
		return err
	}

	store := session.NewStore(cfg)
	game, err := store.LoadOrCreate(pool)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	url := fmt.Sprintf("http://%s", ln.Addr().String())

	srv := server.New(cfg, store, game, pool)
	httpServer := &http.Server{Handler: srv.Handler()}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(ln)
	}()

	fmt.Printf("Bingo for %s — open %s (Ctrl+C to stop)\n", game.Date, url)
	if err := openBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not open browser: %v\n", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		fmt.Printf("\nShutting down (%v)...\n", sig)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
	return nil
}

func openBrowser(url string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name, args = "open", []string{url}
	case "linux":
		name, args = "xdg-open", []string{url}
	case "windows":
		name, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return exec.Command(name, args...).Start()
}
