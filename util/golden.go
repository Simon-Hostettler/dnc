package util

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/cellbuf"
)

var Update = flag.Bool("update", false, "update golden files")

// RenderView passes an ANSI string through a virtual terminal buffer,
// producing a canonical representation. Visually identical strings with
// different escape-code encodings (e.g. \x1b[0m vs \x1b[m) will produce
// identical output.
func RenderView(s string) string {
	w := lipgloss.Width(s)
	h := lipgloss.Height(s)
	if w == 0 || h == 0 {
		return ""
	}
	buf := cellbuf.NewBuffer(w, h)
	cellbuf.SetContent(buf, s)
	out := cellbuf.Render(buf)
	return strings.ReplaceAll(out, "\r\n", "\n")
}

func AssertGolden(t *testing.T, name string, got string) {
	t.Helper()
	got = RenderView(got)
	path := filepath.Join("testdata", name+".golden")
	if *Update {
		os.MkdirAll(filepath.Dir(path), 0o755)
		os.WriteFile(path, []byte(got), 0o644)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden file %s not found (run with -update to create): %v", path, err)
	}
	if got != string(want) {
		t.Errorf("View output differs from golden file %s.\n--- got (len=%d) ---\n%s\n--- want (len=%d) ---\n%s", path, len(got), got, len(want), string(want))
	}
}
