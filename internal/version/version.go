// Package version holds build-time metadata injected via -ldflags "-X".
//
// Values are strings defaulting to "dev"/"unknown" so a plain `go build` or
// `wails3 build` without ldflags still works (tests, local dev). The CI
// workflow overrides them on tag builds.
package version

// These are overwritten by the linker when -ldflags "-X ...=..." is passed.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// Info is the typed shape returned to the frontend by VersionService.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// Get returns the current build metadata as a single value.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
	}
}
