package app

import (
	"strings"
	"testing"

	"wslc-tui-ms/internal/commands"
)

func stripAnsiTest(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		if r == '\x1b' {
			esc = true
			continue
		}
		if esc {
			if r == 'm' {
				esc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func regionForAction(m model, action string) (clickRegion, bool) {
	if m.clickRegions == nil {
		return clickRegion{}, false
	}
	for _, r := range *m.clickRegions {
		if r.action == action {
			return r, true
		}
	}
	return clickRegion{}, false
}

// visibleCell returns the rune at (col, row) in the rendered, ANSI-stripped view.
func visibleCell(view string, col, row int) string {
	lines := strings.Split(view, "\n")
	if row < 0 || row >= len(lines) {
		return ""
	}
	runes := []rune(stripAnsiTest(lines[row]))
	if col < 0 || col >= len(runes) {
		return ""
	}
	return string(runes[col])
}

func lineContains(view string, row int, sub string) bool {
	lines := strings.Split(view, "\n")
	if row < 0 || row >= len(lines) {
		return false
	}
	return strings.Contains(stripAnsiTest(lines[row]), sub)
}

func TestHeaderRegionsAlignWithVisibleText(t *testing.T) {
	for _, w := range []int{90, 120, 150} {
		m := NewModelForTest(w, 30)
		view := m.View()

		cases := []struct {
			action string
			text   string
		}{
			{"tab-commands", "Commands"},
			{"tab-learn", "Learn"},
			{"help", "Help"},
		}
		for _, c := range cases {
			r, ok := regionForAction(m, c.action)
			if !ok {
				t.Fatalf("width=%d: region %q not registered", w, c.action)
			}
			// The region row must actually contain the label text.
			if !lineContains(view, r.y1, c.text) {
				t.Errorf("width=%d: region %q y=%d does not contain %q", w, c.action, r.y1, c.text)
			}
			// The label's visible columns must fall inside the region's X range.
			lines := strings.Split(view, "\n")
			clean := stripAnsiTest(lines[r.y1])
			idx := strings.Index(clean, c.text)
			if idx < 0 {
				t.Fatalf("width=%d: %q not found on row %d", w, c.text, r.y1)
			}
			start := len([]rune(clean[:idx]))
			end := start + len([]rune(c.text)) - 1
			if start < r.x1 || end > r.x2 {
				t.Errorf("width=%d: %q visible cols [%d..%d] not inside region [%d..%d]",
					w, c.text, start, end, r.x1, r.x2)
			}
		}
	}
}

func TestFooterRegionsAlignWithVisibleText(t *testing.T) {
	m := NewModelForTest(120, 30)
	view := m.View()

	// Default (input focused) footer has Esc/Tab/Enter hints.
	cases := []struct {
		action string
		text   string
	}{
		{"esc", "Esc"},
		{"tab", "Tab"},
		{"enter", "Enter"},
	}
	for _, c := range cases {
		r, ok := regionForAction(m, c.action)
		if !ok {
			t.Fatalf("footer region %q not registered", c.action)
		}
		if !lineContains(view, r.y1, c.text) {
			t.Errorf("footer region %q y=%d does not contain %q (row=%q)",
				c.action, r.y1, c.text, stripAnsiTest(strings.Split(view, "\n")[r.y1]))
		}
		lines := strings.Split(view, "\n")
		clean := stripAnsiTest(lines[r.y1])
		idx := strings.Index(clean, c.text)
		if idx < 0 {
			t.Fatalf("%q not found on row %d", c.text, r.y1)
		}
		start := len([]rune(clean[:idx]))
		end := start + len([]rune(c.text)) - 1
		if start < r.x1 || end > r.x2 {
			t.Errorf("footer %q visible cols [%d..%d] not inside region [%d..%d]",
				c.text, start, end, r.x1, r.x2)
		}
	}
}

func TestViewFitsHeight(t *testing.T) {
	for _, dim := range [][2]int{{120, 30}, {100, 24}, {150, 40}} {
		m := NewModelForTest(dim[0], dim[1])
		lines := strings.Count(m.View(), "\n") + 1
		if lines > dim[1] {
			t.Errorf("width=%d height=%d: rendered %d lines (overflow)", dim[0], dim[1], lines)
		}
	}
}

func TestOutputFooterAdvertisesCopyCommand(t *testing.T) {
	m := NewModelForTest(120, 30)
	m.currentView = viewOutput
	m.outputCmd = "wslc container ls"
	m.outputResult = &commands.ExecutionResult{}

	plain := stripAnsiTest(m.View())
	if !strings.Contains(plain, "c Copy command") {
		t.Fatalf("output footer should advertise command copying, got:\n%s", plain)
	}
	if !strings.Contains(plain, "y Copy") {
		t.Fatalf("output footer should retain output copying, got:\n%s", plain)
	}
}
