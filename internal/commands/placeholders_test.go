package commands

import (
	"reflect"
	"testing"
)

func TestExtractPlaceholders(t *testing.T) {
	tests := []struct {
		name string
		full string
		want []string
	}{
		{"none", "wslc container ls", nil},
		{"single", "wslc container run {image}", []string{"image"}},
		{
			"multiple",
			"wslc container run -d --name {name} {image}",
			[]string{"name", "image"},
		},
		{
			"dedup preserves first order",
			"wslc cp {source} {target} {source}",
			[]string{"source", "target"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPlaceholders(tt.full)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ExtractPlaceholders(%q) = %v, want %v", tt.full, got, tt.want)
			}
		})
	}
}

func TestSubstitutePlaceholders(t *testing.T) {
	tests := []struct {
		name   string
		full   string
		values map[string]string
		want   string
	}{
		{
			"all filled",
			"wslc container run -d --name {name} {image}",
			map[string]string{"name": "web", "image": "nginx"},
			"wslc container run -d --name web nginx",
		},
		{
			"missing left unchanged",
			"wslc container run --name {name} {image}",
			map[string]string{"name": "web"},
			"wslc container run --name web {image}",
		},
		{
			"none",
			"wslc container ls",
			map[string]string{},
			"wslc container ls",
		},
		{
			"repeated token",
			"cp {source} {target} && echo {source}",
			map[string]string{"source": "a.txt", "target": "b.txt"},
			"cp a.txt b.txt && echo a.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubstitutePlaceholders(tt.full, tt.values)
			if got != tt.want {
				t.Fatalf("SubstitutePlaceholders(%q) = %q, want %q", tt.full, got, tt.want)
			}
		})
	}
}
