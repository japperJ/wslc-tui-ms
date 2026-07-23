package update

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func VerifyFileSHA256(path, expected string) error {
	want, err := hex.DecodeString(strings.TrimSpace(expected))
	if err != nil || len(want) != sha256.Size {
		return fmt.Errorf("invalid SHA-256 checksum")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), hex.EncodeToString(want)) {
		return fmt.Errorf("SHA-256 checksum mismatch for %s", path)
	}
	return nil
}

func ApplyPortableZip(zipPath, installDir, executable string) error {
	if filepath.IsAbs(executable) || executable == "." || executable == ".." || filepath.Base(executable) != executable {
		return fmt.Errorf("invalid executable name %q", executable)
	}
	parent := filepath.Dir(filepath.Clean(installDir))
	stage, err := os.MkdirTemp(parent, ".wslc-update-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err := extractZip(zipPath, stage); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(stage, executable)); err != nil {
		return fmt.Errorf("portable update is missing %s: %w", executable, err)
	}

	backup := installDir + ".backup"
	if _, err := os.Stat(backup); err == nil {
		return fmt.Errorf("backup path already exists: %s", backup)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(installDir, backup); err != nil {
		return fmt.Errorf("backup current installation: %w", err)
	}
	if err := os.Rename(stage, installDir); err != nil {
		_ = os.Rename(backup, installDir)
		return fmt.Errorf("activate portable update: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("remove installation backup: %w", err)
	}
	return nil
}

func extractZip(path, destination string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()
	root := filepath.Clean(destination) + string(os.PathSeparator)
	for _, entry := range r.File {
		target := filepath.Clean(filepath.Join(destination, filepath.FromSlash(entry.Name)))
		if target != filepath.Clean(destination) && !strings.HasPrefix(target, root) {
			return fmt.Errorf("zip entry escapes destination: %s", entry.Name)
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		in, err := entry.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o700)
		if err == nil {
			_, err = io.Copy(out, in)
			if closeErr := out.Close(); err == nil {
				err = closeErr
			}
		}
		in.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
