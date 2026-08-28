package version

import "strings"

var (
	// Version is injected at build time via -ldflags
	Version = "1.2.0"
	// Commit is injected at build time via -ldflags
	Commit = "dev"
	// BuildDate is injected at build time via -ldflags
	BuildDate = "unknown"
)

// FormattedVersion returns the clean semantic version with a single 'v' prefix.
func FormattedVersion() string {
	return "v" + strings.TrimPrefix(Version, "v")
}
