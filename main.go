package main

import "bingo/cmd"

// CMD to build: go build -ldflags "-X bingo/internal/version.Version=1.2.3 -X bingo/internal/version.Updated=2026-07-29"

func main() {
	cmd.Execute()
}
