package app

import (
	"strings"
	"testing"
)

func TestLearnContentCoversNewWSLCCommands(t *testing.T) {
	tests := []struct {
		topic    string
		commands []string
	}{
		{"Container Operations", []string{"wslc container cp", "--archive"}},
		{"Registry", []string{"wslc login", "wslc logout", "--password-stdin"}},
	}

	for _, test := range tests {
		t.Run(test.topic, func(t *testing.T) {
			content := (model{}).getLearnContent(test.topic)
			for _, command := range test.commands {
				if !strings.Contains(content, command) {
					t.Errorf("learn topic %q does not mention %q", test.topic, command)
				}
			}
		})
	}
}
