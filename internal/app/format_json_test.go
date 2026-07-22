package app

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatJSONSingleObject(t *testing.T) {
	in := `{"a":1,"b":"x"}`
	got := formatJSON(in)
	if !strings.Contains(got, "\n") {
		t.Fatalf("expected indented multi-line output, got %q", got)
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(got), &v); err != nil {
		t.Fatalf("formatted single object is not valid JSON: %v\n%s", err, got)
	}
}

func TestFormatJSONLinesMultiple(t *testing.T) {
	in := `{"ID":"43f0fc26d1d6","Names":"web"}
{"ID":"a6f098a80160","Names":"n8n"}`
	got := formatJSON(in)

	if !strings.Contains(got, "43f0fc26d1d6") || !strings.Contains(got, "a6f098a80160") {
		t.Fatalf("both objects should be present:\n%s", got)
	}
	// Each object should be indented (pretty-printed), so both IDs appear on
	// their own indented lines rather than a single unformatted blob.
	if strings.Count(got, "\n  \"") < 2 {
		t.Errorf("expected each object indented, got:\n%s", got)
	}
	// Objects must be separated so they don't merge into one line.
	if strings.Contains(got, `}{`) || strings.Contains(got, "}\n{") {
		t.Errorf("adjacent objects should be blank-line separated, got:\n%s", got)
	}
}

func TestFormatJSONNonJSONPassthrough(t *testing.T) {
	in := "CONTAINER ID   IMAGE\n43f0fc26d1d6   nginx"
	got := formatJSON(in)
	if got != in {
		t.Errorf("non-JSON input should pass through unchanged.\nin:  %q\ngot: %q", in, got)
	}
}

func TestFormatJSONLinesWithBlankLines(t *testing.T) {
	in := "\n{\"a\":1}\n\n{\"b\":2}\n"
	got := formatJSON(in)
	if !strings.Contains(got, "\"a\"") || !strings.Contains(got, "\"b\"") {
		t.Fatalf("expected both objects, got:\n%s", got)
	}
}
