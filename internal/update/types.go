package update

import "context"

type Channel string

const (
	Stable Channel = "stable"
	Beta   Channel = "beta"
)

type Asset struct {
	Name   string `json:"name"`
	URL    string `json:"url,omitempty"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"sizeBytes"`
}
type Release struct {
	Tag, Notes, URL string
	Prerelease      bool
	Assets          []Asset
}
type Policy struct{ MinimumSupportedVersion, Message string }
type Decision struct {
	Available bool
	Mandatory bool
	// RequiredVersion carries the minimum supported version from policy when the
	// running build must be upgraded. It is set even when no installable release
	// exists yet, so callers can report why the update is required without
	// treating an empty Version as actionable.
	RequiredVersion              string
	Version, Channel, Notes, URL string
	Asset                        Asset
	Release                      Release
}

type ReleaseClient interface {
	Releases(context.Context) ([]Release, error)
	Policy(context.Context) (Policy, error)
	Checksums(context.Context, Release) (map[string]Asset, error)
}
