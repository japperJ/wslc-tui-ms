package update

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunInstallerWithTimeoutReportsTimeout(t *testing.T) {
	err := runInstallerWithTimeout(20*time.Millisecond, os.Args[0], "-test.run=TestInstallerHelperProcess")
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("runInstallerWithTimeout error = %v, want timeout", err)
	}
}

func TestAcquireUpdateLockRejectsConcurrentUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".wslc-update.lock")
	release, err := acquireUpdateLock(path)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := acquireUpdateLock(path); err == nil {
		t.Fatal("second update lock acquisition should fail")
	}
}

func TestAcquireUpdateLockRemovesStaleOwner(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".wslc-update.lock")
	if err := os.WriteFile(path, []byte("999999999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	release, err := acquireUpdateLock(path)
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestInstallerHelperProcess(t *testing.T) {
	if !strings.Contains(strings.Join(os.Args, " "), "-test.run=TestInstallerHelperProcess") {
		return
	}
	time.Sleep(time.Second)
}

func TestTempAssetPatternPreservesInstallerExtension(t *testing.T) {
	tests := []struct {
		distribution string
		assetName    string
		want         string
	}{
		{"msi", "wslc-tui-v1.2.8-beta.1-windows-amd64.msi", "wslc-update-*.msi"},
		{"installer", "wslc-tui-v1.2.8-beta.1-windows-amd64.msi", "wslc-update-*.msi"},
		{"exe", "wslc-tui-v1.2.8-beta.1-windows-amd64.exe", "wslc-update-*.exe"},
		{"portable", "wslc-tui-v1.2.8-beta.1-windows-amd64-portable.zip", "wslc-update-*.zip"},
	}
	for _, tt := range tests {
		t.Run(tt.distribution, func(t *testing.T) {
			got, err := tempAssetPattern(Handoff{Distribution: tt.distribution, AssetName: tt.assetName})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("temp pattern = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyPortableZipReplacesInstallAndRemovesBackup(t *testing.T) {
	installDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(installDir, "wslc-tui.exe"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(t.TempDir(), "update.zip")
	writeUpdateZip(t, zipPath, "new")

	if err := ApplyPortableZip(zipPath, installDir, "wslc-tui.exe"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(installDir, "wslc-tui.exe"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("installed executable = %q, want new", got)
	}
	if _, err := os.Stat(installDir + ".backup"); !os.IsNotExist(err) {
		t.Fatalf("backup directory still exists: %v", err)
	}
}

func TestApplyPortableZipRejectsMissingExecutableWithoutChangingInstall(t *testing.T) {
	installDir := t.TempDir()
	original := filepath.Join(installDir, "wslc-tui.exe")
	if err := os.WriteFile(original, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(t.TempDir(), "update.zip")
	writeUpdateZipEntry(t, zipPath, "README.txt", []byte("not an app"))

	if err := ApplyPortableZip(zipPath, installDir, "wslc-tui.exe"); err == nil {
		t.Fatal("missing executable should fail")
	}
	got, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old" {
		t.Fatalf("install was changed after rejection: %q", got)
	}
}

func TestVerifyFileSHA256(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload")
	contents := []byte("verified payload")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(contents)
	if err := VerifyFileSHA256(path, hex.EncodeToString(sum[:])); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFileSHA256(path, "bad"); err == nil {
		t.Fatal("invalid checksum should fail")
	}
}

func TestDownloadAndInstallPortableZipVerifiesBeforeReplacement(t *testing.T) {
	installDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(installDir, "wslc-tui.exe"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(t.TempDir(), "update.zip")
	writeUpdateZip(t, zipPath, "new")
	payload, err := os.ReadFile(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	handoff := Handoff{
		SchemaVersion: HandoffSchemaVersion,
		AttemptID:     "attempt-download",
		Distribution:  "portable",
		AssetURL:      server.URL,
		AssetName:     "wslc-tui-v2.0.0-windows-amd64-portable.zip",
		SHA256:        hex.EncodeToString(sum[:]),
		InstallDir:    installDir,
		CurrentExe:    filepath.Join(installDir, "wslc-tui.exe"),
		TargetVersion: "v2.0.0",
		ResultPath:    filepath.Join(t.TempDir(), "result.json"),
		ParentPID:     os.Getpid(),
	}
	if err := DownloadAndInstall(context.Background(), server.Client(), handoff); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(installDir, "wslc-tui.exe"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Fatalf("installed executable = %q, want new", got)
	}
}

func writeUpdateZip(t *testing.T, path, executable string) {
	t.Helper()
	writeUpdateZipEntry(t, path, "wslc-tui.exe", []byte(executable))
}

func writeUpdateZipEntry(t *testing.T, path, name string, contents []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	entry, err := w.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}
