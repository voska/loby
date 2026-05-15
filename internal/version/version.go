// Package version exposes build information embedded at link time.
package version

import (
	"runtime"
	"runtime/debug"
)

// These are overwritten by -ldflags at build time. See Makefile.
var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

// Info is the structured build information surfaced by `loby version --json`.
type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Go      string `json:"go"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// Get returns the current build info, filling in module-derived defaults when
// the binary was built without ldflag injection (e.g. `go install`).
func Get() Info {
	v, c := Version, Commit
	if v == "dev" {
		if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			v = bi.Main.Version
		}
	}
	if c == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			for _, s := range bi.Settings {
				if s.Key == "vcs.revision" && len(s.Value) >= 12 {
					c = s.Value[:12]
				}
			}
		}
	}
	return Info{
		Version: v,
		Commit:  c,
		Date:    Date,
		Go:      runtime.Version(),
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}
}
