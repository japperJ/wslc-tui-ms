package update

import (
	"context"
	"fmt"
	"time"
	"wslc-tui-ms/internal/settings"
)

const Cooldown = 24 * time.Hour

type Service struct {
	Client                       ReleaseClient
	Store                        settings.Store
	Now                          func() time.Time
	Distribution, CurrentVersion string
}

func (s Service) Check(ctx context.Context, channel Channel, manual bool) (Decision, error) {
	now := time.Now()
	if s.Now != nil {
		now = s.Now()
	}
	state, err := s.Store.Load()
	if err != nil && !manual {
		return Decision{}, nil
	}
	if err != nil {
		return Decision{}, err
	}
	if !manual && state.LastCheck != "" {
		if t, e := time.Parse(time.RFC3339Nano, state.LastCheck); e == nil && now.Sub(t) < Cooldown {
			return Decision{}, nil
		}
	}
	releases, err := s.Client.Releases(ctx)
	if err != nil {
		if manual {
			return Decision{}, err
		}
		return Decision{}, nil
	}
	state.LastCheck = now.Format(time.RFC3339Nano)
	if err := s.Store.Save(state); err != nil && manual {
		return Decision{}, err
	}
	policy, err := s.Client.Policy(ctx)
	if err != nil {
		if manual {
			return Decision{}, err
		}
		return Decision{}, nil
	}
	if _, err := parseVersion(policy.MinimumSupportedVersion); err != nil {
		if manual {
			return Decision{}, fmt.Errorf("invalid minimum supported version: %w", err)
		}
		return Decision{}, nil
	}
	minimum := newer(policy.MinimumSupportedVersion, s.CurrentVersion)
	selected, err := selectRelease(releases, channel)
	if err != nil {
		return Decision{}, err
	}
	if selected == nil {
		return Decision{Mandatory: minimum}, nil
	}
	if !newer(selected.Tag, s.CurrentVersion) && !minimum {
		return Decision{}, nil
	}
	if !minimum && state.Deferred == selected.Tag {
		return Decision{}, nil
	}
	checksums, err := s.Client.Checksums(ctx, *selected)
	if err != nil {
		if manual {
			return Decision{}, err
		}
		return Decision{}, nil
	}
	assetName := assetNameFor(s.Distribution, selected.Tag)
	asset, ok := findAsset(selected.Assets, assetName)
	if !ok {
		return Decision{}, fmt.Errorf("release %s is missing %s", selected.Tag, assetName)
	}
	if checksum, ok := checksums[assetName]; ok {
		asset.SHA256, asset.Size = checksum.SHA256, checksum.Size
	} else {
		return Decision{}, fmt.Errorf("checksum manifest is missing %s", assetName)
	}
	return Decision{Available: true, Mandatory: minimum, Version: selected.Tag, Channel: string(channel), Notes: selected.Notes, URL: selected.URL, Asset: asset, Release: *selected}, nil
}
func (s Service) Defer(version string) error {
	state, err := s.Store.Load()
	if err != nil {
		return err
	}
	state.Deferred = version
	return s.Store.Save(state)
}
func selectRelease(releases []Release, channel Channel) (*Release, error) {
	var best *Release
	for i := range releases {
		r := &releases[i]
		if _, err := parseVersion(r.Tag); err != nil {
			continue
		}
		if channel == Stable && r.Prerelease {
			continue
		}
		if channel != Stable && channel != Beta {
			return nil, fmt.Errorf("unknown channel %q", channel)
		}
		if best == nil {
			best = r
			continue
		}
		a, _ := parseVersion(r.Tag)
		b, _ := parseVersion(best.Tag)
		if compare(a, b) > 0 {
			best = r
		}
	}
	return best, nil
}
func assetNameFor(distribution, tag string) string {
	switch distribution {
	case "installer", "msi":
		return "wslc-tui-" + tag + "-windows-amd64.msi"
	case "exe":
		return "wslc-tui-" + tag + "-windows-amd64.exe"
	default:
		return "wslc-tui-" + tag + "-windows-amd64-portable.zip"
	}
}
func findAsset(assets []Asset, name string) (Asset, bool) {
	for _, a := range assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}
