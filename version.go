package rope

import "fmt"

// Version is the current semantic version of the library.
const Version = "0.1.0"

// Variables set by linker flags during build.
var (
	CommitHash = "unknown"
	BuildTime  = "unknown"
)

// BuildInfo holds information about the library build.
type BuildInfo struct {
	Version    string
	CommitHash string
	BuildTime  string
}

// GetBuildInfo returns the current build information.
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:    Version,
		CommitHash: CommitHash,
		BuildTime:  BuildTime,
	}
}

// String returns a formatted version string.
func (b BuildInfo) String() string {
	return fmt.Sprintf("Rope v%s (%s) built at %s", b.Version, b.CommitHash, b.BuildTime)
}
