package update

import "context"

type Channel string

const (
	Stable Channel = "stable"
	Beta   Channel = "beta"
)

type Asset struct {
	Name, URL, SHA256 string
	Size              int64
}
type Release struct {
	Tag, Notes, URL string
	Prerelease      bool
	Assets          []Asset
}
type Policy struct{ MinimumSupportedVersion, Message string }
type Decision struct {
	Available                    bool
	Mandatory                    bool
	Version, Channel, Notes, URL string
	Asset                        Asset
	Release                      Release
}

type ReleaseClient interface {
	Releases(context.Context) ([]Release, error)
	Policy(context.Context) (Policy, error)
	Checksums(context.Context, Release) (map[string]Asset, error)
}
