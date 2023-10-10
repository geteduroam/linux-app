package version

import (
	"fmt"
	"runtime/debug"
)

// version is the current version
const version = "0.1"

// isReleased sets whether or not the current version is released yet
const isReleased = true

// Get gets the version in the following order:
// - Gets a release version if it detects it is a release
// - Gets the commit using debug info
// - Returns a default
func Get() string {
	if isReleased {
		return fmt.Sprintf("Version: %s", version)
	}
	if dbg, ok := debug.ReadBuildInfo(); ok {
		for _, s := range dbg.Settings {
			if s.Key == "vcs.revision" {
				return fmt.Sprintf("Dev version: %s with commit: %s", version, s.Value)
			}
		}
	}
	return fmt.Sprintf("Dev version: %s", version)
}
