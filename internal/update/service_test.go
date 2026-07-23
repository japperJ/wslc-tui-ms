package update

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
	"wslc-tui-ms/internal/settings"
)

type fakeClient struct {
	releases  []Release
	policy    Policy
	checksums map[string]Asset
	calls     int
	err       error
}

func (f *fakeClient) Releases(context.Context) ([]Release, error) {
	f.calls++
	return f.releases, f.err
}
func (f *fakeClient) Policy(context.Context) (Policy, error) { return f.policy, nil }
func (f *fakeClient) Checksums(context.Context, Release) (map[string]Asset, error) {
	return f.checksums, nil
}
func testService(t *testing.T, f *fakeClient) Service {
	t.Helper()
	return Service{Client: f, Store: settings.NewStore(filepath.Join(t.TempDir(), "settings.json")), CurrentVersion: "v1.0.0", Distribution: "portable", Now: func() time.Time { return time.Unix(100, 0) }}
}
func release(tag string, beta bool) Release {
	return Release{Tag: tag, Prerelease: beta, Notes: "release notes", Assets: []Asset{{Name: "wslc-tui-" + tag + "-windows-amd64-portable.zip", URL: "https://download"}}}
}
func checksum(tag string) map[string]Asset {
	return map[string]Asset{"wslc-tui-" + tag + "-windows-amd64-portable.zip": {SHA256: "hash", Size: 42}}
}

func TestServiceFiltersStableAndBeta(t *testing.T) {
	f := &fakeClient{releases: []Release{release("v2.0.0", false), release("v3.0.0-beta.1", true)}, policy: Policy{MinimumSupportedVersion: "0.0.0"}, checksums: checksum("v2.0.0")}
	s := testService(t, f)
	d, err := s.Check(context.Background(), Stable, true)
	if err != nil || d.Version != "v2.0.0" {
		t.Fatalf("stable=%+v err=%v", d, err)
	}
	f.checksums = checksum("v3.0.0-beta.1")
	d, err = s.Check(context.Background(), Beta, true)
	if err != nil || d.Version != "v3.0.0-beta.1" {
		t.Fatalf("beta=%+v err=%v", d, err)
	}
}
func TestServiceSkipsMalformedAndSupportsPrereleasePromotion(t *testing.T) {
	f := &fakeClient{releases: []Release{release("not-semver", false), release("v2.0.0-rc.1", true), release("v1.5.0", false)}, policy: Policy{MinimumSupportedVersion: "0.0.0"}, checksums: checksum("v2.0.0-rc.1")}
	d, err := testService(t, f).Check(context.Background(), Beta, true)
	if err != nil || d.Version != "v2.0.0-rc.1" {
		t.Fatalf("decision=%+v err=%v", d, err)
	}
}
func TestServicePolicyAndDeferralPersist(t *testing.T) {
	f := &fakeClient{releases: []Release{release("v1.1.0", false)}, policy: Policy{MinimumSupportedVersion: "v1.1.0"}, checksums: checksum("v1.1.0")}
	s := testService(t, f)
	d, err := s.Check(context.Background(), Stable, true)
	if err != nil || !d.Mandatory {
		t.Fatalf("mandatory=%+v err=%v", d, err)
	}
	if err := s.Defer("v1.1.0"); err != nil {
		t.Fatal(err)
	}
	d, err = s.Check(context.Background(), Stable, true)
	if err != nil || !d.Available {
		t.Fatalf("policy must ignore deferral: %+v %v", d, err)
	}
}
func TestServiceCooldownAcrossInstancesAndManualBypass(t *testing.T) {
	f := &fakeClient{releases: []Release{release("v2.0.0", false)}, policy: Policy{MinimumSupportedVersion: "0.0.0"}, checksums: checksum("v2.0.0")}
	s := testService(t, f)
	if _, err := s.Check(context.Background(), Stable, false); err != nil {
		t.Fatal(err)
	}
	s2 := s
	if _, err := s2.Check(context.Background(), Stable, false); err != nil {
		t.Fatal(err)
	}
	if f.calls != 1 {
		t.Fatalf("automatic calls=%d", f.calls)
	}
	if _, err := s2.Check(context.Background(), Stable, true); err != nil {
		t.Fatal(err)
	}
	if f.calls != 2 {
		t.Fatalf("manual calls=%d", f.calls)
	}
}
func TestServiceErrorsAreSilentOnlyInBackground(t *testing.T) {
	f := &fakeClient{err: errors.New("offline")}
	s := testService(t, f)
	if _, err := s.Check(context.Background(), Stable, false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Check(context.Background(), Stable, true); err == nil {
		t.Fatal("manual check should expose error")
	}
}
func TestServicePreservesAssetMetadata(t *testing.T) {
	f := &fakeClient{releases: []Release{release("v2.0.0", false)}, policy: Policy{MinimumSupportedVersion: "0.0.0"}, checksums: checksum("v2.0.0")}
	d, err := testService(t, f).Check(context.Background(), Stable, true)
	if err != nil {
		t.Fatal(err)
	}
	if d.Asset.SHA256 != "hash" || d.Asset.Size != 42 || d.Asset.URL != "https://download" || d.Notes != "release notes" {
		t.Fatalf("metadata lost: %+v", d)
	}
}
