package update

import (
	"path/filepath"
	"testing"
)

func TestHandoffRoundTripIsAtomic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handoff.json")
	want := Handoff{
		SchemaVersion: 1,
		AttemptID:     "attempt-123",
		Distribution:  "portable",
		AssetURL:      "https://example.invalid/update.zip",
		AssetName:     "wslc-tui-v1.2.3-windows-amd64-portable.zip",
		SHA256:        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		InstallDir:    filepath.Dir(path),
		CurrentExe:    filepath.Join(filepath.Dir(path), "wslc-tui.exe"),
		TargetVersion: "v1.2.3",
		ResultPath:    filepath.Join(filepath.Dir(path), "result.json"),
		RelaunchArgs:  []string{"--demo", "value with spaces"},
		ParentPID:     123,
	}
	if err := WriteHandoff(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := ReadHandoff(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.AttemptID != want.AttemptID || got.AssetURL != want.AssetURL || len(got.RelaunchArgs) != 2 {
		t.Fatalf("handoff round trip mismatch: %#v", got)
	}
}

func TestReadHandoffRejectsUnsupportedDistribution(t *testing.T) {
	path := filepath.Join(t.TempDir(), "handoff.json")
	if err := WriteHandoff(path, Handoff{SchemaVersion: 1, AttemptID: "a", Distribution: "unknown"}); err == nil {
		t.Fatal("unsupported distribution should fail")
	}
}

func TestInstallerDistributionIsAcceptedAsMSI(t *testing.T) {
	handoff := Handoff{
		SchemaVersion: HandoffSchemaVersion,
		AttemptID:     "installer-attempt",
		Distribution:  "installer",
		AssetURL:      "https://example.invalid/update.msi",
		AssetName:     "wslc-tui-v1.2.4-beta.2-windows-amd64.msi",
		SHA256:        "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		InstallDir:    `C:\Users\test\AppData\Local\wslc-tui-ms`,
		CurrentExe:    `C:\Users\test\AppData\Local\wslc-tui-ms\wslc-tui.exe`,
		TargetVersion: "v1.2.4-beta.2",
		ResultPath:    `C:\Users\test\AppData\Roaming\wslc-tui-ms\update-result.json`,
		ParentPID:     123,
	}
	if err := handoff.Validate(); err != nil {
		t.Fatal(err)
	}
}
