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
	"strconv"
	"strings"
	"time"
)

func DownloadAndInstall(ctx context.Context, client *http.Client, handoff Handoff) error {
	if err := handoff.Validate(); err != nil {
		return err
	}
	releaseLock, err := acquireUpdateLock(filepath.Join(handoff.InstallDir, ".wslc-update.lock"))
	if err != nil {
		return err
	}
	defer releaseLock()
	tempPattern, err := tempAssetPattern(handoff)
	if err != nil {
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
	tmp, err := os.CreateTemp("", tempPattern)
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
		installerLog := filepath.Join(filepath.Dir(handoff.ResultPath), "installer.log")
		return runInstaller("msiexec.exe", "/i", tmpPath, "/quiet", "/norestart", "/l*v", installerLog)
	case "exe":
		return runInstaller(tmpPath, "/quiet", "/norestart")
	default:
		return fmt.Errorf("unsupported distribution %q", handoff.Distribution)
	}
}

func tempAssetPattern(handoff Handoff) (string, error) {
	ext := filepath.Ext(handoff.AssetName)
	want := map[string]string{
		"portable":  ".zip",
		"msi":       ".msi",
		"installer": ".msi",
		"exe":       ".exe",
	}[handoff.Distribution]
	if want == "" || !strings.EqualFold(ext, want) {
		return "", fmt.Errorf("asset %q has unexpected extension for %s distribution", handoff.AssetName, handoff.Distribution)
	}
	return "wslc-update-*" + want, nil
}

func runInstaller(program string, args ...string) error {
	if strings.EqualFold(filepath.Base(program), "msiexec.exe") {
		// Killing the msiexec client does not cancel Windows Installer's service
		// transaction; it can continue in the background and install later.
		if err := exec.Command(program, args...).Run(); err != nil {
			return fmt.Errorf("installer failed (exit code %d): %w", exitCode(err), err)
		}
		return nil
	}
	return runInstallerWithTimeout(5*time.Minute, program, args...)
}

func runInstallerWithTimeout(timeout time.Duration, program string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := exec.CommandContext(ctx, program, args...).Run()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("installer timed out after %s", timeout)
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1618 {
			return fmt.Errorf("another Windows Installer operation is already in progress (exit code 1618)")
		}
		return fmt.Errorf("installer failed (exit code %d): %w", exitCode(err), err)
	}
	return nil
}

func exitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

func acquireUpdateLock(path string) (func(), error) {
	lock, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			if b, readErr := os.ReadFile(path); readErr == nil {
				pid, parseErr := strconv.Atoi(strings.TrimSpace(string(b)))
				if parseErr == nil && !processIsRunning(pid) {
					_ = os.Remove(path)
					return acquireUpdateLock(path)
				}
			}
			return nil, fmt.Errorf("another update is already in progress")
		}
		return nil, fmt.Errorf("create update lock: %w", err)
	}
	if _, err := fmt.Fprintf(lock, "%d\n", os.Getpid()); err != nil {
		_ = lock.Close()
		_ = os.Remove(path)
		return nil, fmt.Errorf("write update lock: %w", err)
	}
	if err := lock.Close(); err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("close update lock: %w", err)
	}
	return func() { _ = os.Remove(path) }, nil
}
