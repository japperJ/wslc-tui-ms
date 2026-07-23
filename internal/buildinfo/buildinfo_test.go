package buildinfo

import "testing"

func TestDevelopmentDefaults(t *testing.T) {
	if got := Current(); got.Version != "dev" || got.Commit != "unknown" || got.BuildDate != "unknown" || got.Channel != "development" || got.Distribution != "development" {
		t.Fatalf("unexpected development metadata: %+v", got)
	}
}

func TestInjectedReleaseMetadata(t *testing.T) {
	old := Current()
	t.Cleanup(func() {
		Version, Commit, BuildDate, Channel, Distribution = old.Version, old.Commit, old.BuildDate, old.Channel, old.Distribution
	})

	Version = "v1.2.3"
	Commit = "abc123"
	BuildDate = "2026-07-23T00:00:00Z"
	Channel = "Beta"
	Distribution = "installer"

	got := Current()
	if got.Version != "v1.2.3" || got.Commit != "abc123" || got.BuildDate != "2026-07-23T00:00:00Z" || got.Channel != "Beta" || got.Distribution != "installer" {
		t.Fatalf("injected metadata was not preserved: %+v", got)
	}
	if got.String() != "wslc-tui v1.2.3 (channel=Beta, distribution=installer, commit=abc123, buildDate=2026-07-23T00:00:00Z)" {
		t.Fatalf("unexpected metadata string: %s", got.String())
	}
}
