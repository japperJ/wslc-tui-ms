package update

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const HandoffSchemaVersion = 1

type Result struct {
	AttemptID string `json:"attemptId"`
	Version   string `json:"version"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

type Handoff struct {
	SchemaVersion int      `json:"schemaVersion"`
	AttemptID     string   `json:"attemptId"`
	Distribution  string   `json:"distribution"`
	AssetURL      string   `json:"assetUrl"`
	AssetName     string   `json:"assetName"`
	SHA256        string   `json:"sha256"`
	AssetSize     int64    `json:"assetSize,omitempty"`
	InstallDir    string   `json:"installDir"`
	CurrentExe    string   `json:"currentExe"`
	TargetVersion string   `json:"targetVersion"`
	ResultPath    string   `json:"resultPath"`
	RelaunchArgs  []string `json:"relaunchArgs"`
	WorkingDir    string   `json:"workingDir,omitempty"`
	ParentPID     int      `json:"parentPid"`
}

func (h Handoff) Validate() error {
	if h.SchemaVersion != HandoffSchemaVersion {
		return fmt.Errorf("unsupported handoff schema version %d", h.SchemaVersion)
	}
	if h.AttemptID == "" || h.TargetVersion == "" || h.AssetURL == "" || h.AssetName == "" || h.SHA256 == "" {
		return fmt.Errorf("handoff is missing required fields")
	}
	if h.Distribution != "portable" && h.Distribution != "msi" && h.Distribution != "installer" && h.Distribution != "exe" {
		return fmt.Errorf("unsupported distribution %q", h.Distribution)
	}
	if h.InstallDir == "" || h.CurrentExe == "" || h.ResultPath == "" || h.ParentPID <= 0 {
		return fmt.Errorf("handoff is missing paths")
	}
	expected := assetNameFor(h.Distribution, h.TargetVersion)
	if h.AssetName != expected {
		return fmt.Errorf("asset %q does not match %s distribution", h.AssetName, h.Distribution)
	}
	return nil
}

func WriteHandoff(path string, handoff Handoff) error {
	if err := handoff.Validate(); err != nil {
		return err
	}
	return writeJSONAtomic(path, handoff)
}

func ReadHandoff(path string) (Handoff, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Handoff{}, err
	}
	var handoff Handoff
	if err := json.Unmarshal(b, &handoff); err != nil {
		return Handoff{}, err
	}
	if err := handoff.Validate(); err != nil {
		return Handoff{}, err
	}
	return handoff, nil
}

func WriteResult(path string, result Result) error {
	if result.AttemptID == "" || result.Status == "" {
		return fmt.Errorf("result is missing required fields")
	}
	return writeJSONAtomic(path, result)
}

func ConsumeResult(path string) (Result, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Result{}, err
	}
	var result Result
	if err := json.Unmarshal(b, &result); err != nil {
		return Result{}, err
	}
	if result.AttemptID == "" || result.Status == "" {
		return Result{}, fmt.Errorf("result is missing required fields")
	}
	if err := os.Remove(path); err != nil {
		return Result{}, err
	}
	return result, nil
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".wslc-atomic-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
