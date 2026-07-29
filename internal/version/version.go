package version

// Overridable at build time, e.g.:
//
//	go build -ldflags "-X bingo/internal/version.Version=1.2.3 -X bingo/internal/version.Updated=2026-07-29"
var (
	Version = "0.1.0"
	Updated = "2026-07-29"
)
