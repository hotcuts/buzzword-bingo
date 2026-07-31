# Buzzword Bingo

Another meeting. Another circle-back. Perfect alignment.

Buzzword Bingo is a board in your browser — a hobby that looks like note-taking. Mark cells as the jargon rolls in. Five in a row and you win. Spiritually, if not professionally. Choose a **daily** or **weekly** (ISO week, Monday start) board in onboarding or settings.

**Supported platforms:** macOS Apple Silicon, Linux amd64, Windows amd64.

## Install

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/hotcuts/buzzword-bingo/main/scripts/install.sh | bash
```

Installs to `~/.local/bin/bingo`. Restart your shell if `bingo` is not found.

### Windows

```powershell
irm https://raw.githubusercontent.com/hotcuts/buzzword-bingo/main/scripts/install.ps1 | iex
```

Installs to `%LOCALAPPDATA%\bingo\bingo.exe` and adds that folder to your user PATH. Open a new terminal if `bingo` is not found.

### Pin a version

```bash
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/hotcuts/buzzword-bingo/main/scripts/install.sh | bash
```

```powershell
$env:VERSION = "v0.1.0"; irm https://raw.githubusercontent.com/hotcuts/buzzword-bingo/main/scripts/install.ps1 | iex
```

## Play

```bash
bingo play
```

Opens the current board in your browser. First launch (and after a full reset) walks through look, name, reset period (daily or ISO weekly), and buzzwords. Settings cover looks, light/dark mode, your name, reset period, and custom terms. If a newer release is available, play may print a one-line hint to run `bingo update` — it never blocks the game.

Set the reset period from the CLI:

```bash
bingo set period weekly   # ISO calendar week (Monday start)
bingo set period daily
bingo set period          # show current period
```

## Update

```bash
bingo update
```

Re-runs the platform install script and replaces your binary with the latest release.
