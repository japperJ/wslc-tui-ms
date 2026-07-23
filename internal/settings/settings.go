package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Channel string

const (
	Stable Channel = "stable"
	Beta   Channel = "beta"
)

type Settings struct {
	Channel   Channel `json:"channel"`
	LastCheck string  `json:"lastCheck,omitempty"`
	Deferred  string  `json:"deferredVersion,omitempty"`
}

type Store struct{ Path string }

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wslc-tui-ms", "settings.json"), nil
}

func NewStore(path string) Store { return Store{Path: path} }

func (s Store) Load() (Settings, error) {
	result := Settings{Channel: Stable}
	b, err := os.ReadFile(s.Path)
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return result, err
	}
	if result.Channel != Stable && result.Channel != Beta {
		result.Channel = Stable
	}
	return result, nil
}

func (s Store) Save(value Settings) error {
	if value.Channel != Stable && value.Channel != Beta {
		value.Channel = Stable
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.Path)
}
