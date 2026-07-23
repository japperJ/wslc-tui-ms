package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func DownloadAndInstall(ctx context.Context, client *http.Client, handoff Handoff) error {
	if err := handoff.Validate(); err != nil {
		return err
	}
	u, err := url.Parse(handoff.AssetURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("invalid update asset URL")
	}
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, handoff.AssetURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download returned %s", resp.Status)
	}
	if handoff.AssetSize > 0 && resp.ContentLength >= 0 && resp.ContentLength != handoff.AssetSize {
		return fmt.Errorf("download size %d does not match expected size %d", resp.ContentLength, handoff.AssetSize)
	}
	tmp, err := os.CreateTemp("", "wslc-update-*.download")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	maxBytes := int64(2 << 30)
	if handoff.AssetSize > 0 {
		maxBytes = handoff.AssetSize + 1
	}
	n, copyErr := io.Copy(tmp, io.LimitReader(resp.Body, maxBytes))
	closeErr := tmp.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if handoff.AssetSize > 0 && n != handoff.AssetSize {
		return fmt.Errorf("download size %d does not match expected size %d", n, handoff.AssetSize)
	}
	if err := VerifyFileSHA256(tmpPath, handoff.SHA256); err != nil {
		return err
	}
	switch handoff.Distribution {
	case "portable":
		return ApplyPortableZip(tmpPath, handoff.InstallDir, filepath.Base(handoff.CurrentExe))
	case "msi", "installer":
		return runInstaller("msiexec.exe", "/i", tmpPath, "/quiet", "/norestart")
	case "exe":
		return runInstaller(tmpPath, "/quiet", "/norestart")
	default:
		return fmt.Errorf("unsupported distribution %q", handoff.Distribution)
	}
}

func runInstaller(program string, args ...string) error {
	if err := exec.Command(program, args...).Run(); err != nil {
		return fmt.Errorf("installer failed: %w", err)
	}
	return nil
}
