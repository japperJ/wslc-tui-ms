// Package buildinfo exposes metadata embedded into release binaries.
package buildinfo

// These variables are the stable linker contract used by release builds.
// Development defaults are deliberately human-readable and never masquerade
// as a tagged release.
var (
	Version      = "dev"
	Commit       = "unknown"
	BuildDate    = "unknown"
	Channel      = "development"
	Distribution = "development"
)

// Info is the metadata shown by the command-line version path.
type Info struct {
	Version      string
	Commit       string
	BuildDate    string
	Channel      string
	Distribution string
}

// Current returns a snapshot of the embedded build metadata.
func Current() Info {
	return Info{
		Version:      Version,
		Commit:       Commit,
		BuildDate:    BuildDate,
		Channel:      Channel,
		Distribution: Distribution,
	}
}

// String returns a stable, machine-readable version line for release smoke
// tests and a concise human-readable output for users.
func (i Info) String() string {
	return "wslc-tui " + i.Version + " (channel=" + i.Channel + ", distribution=" + i.Distribution + ", commit=" + i.Commit + ", buildDate=" + i.BuildDate + ")"
}
