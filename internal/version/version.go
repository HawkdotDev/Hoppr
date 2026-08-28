package version

var (
	// Version is injected at build time via -ldflags
	Version = "1.0.0"
	// Commit is injected at build time via -ldflags
	Commit = "dev"
	// BuildDate is injected at build time via -ldflags
	BuildDate = "unknown"
)
