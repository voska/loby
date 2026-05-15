package output

import (
	"bytes"
	"strings"
	"testing"
)

func newTestWriter(opts Options) (*Writer, *bytes.Buffer, *bytes.Buffer) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	opts.Stdout = stdout
	opts.Stderr = stderr
	return New(opts), stdout, stderr
}

func TestRender_JSON(t *testing.T) {
	w, out, _ := newTestWriter(Options{JSONFlag: true})
	if err := w.Render(map[string]any{"id": "psc_123", "to": "Alice"}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, `"id": "psc_123"`) || !strings.Contains(got, `"to": "Alice"`) {
		t.Fatalf("json out = %q", got)
	}
}

func TestRender_Plain_Slice(t *testing.T) {
	w, out, _ := newTestWriter(Options{PlainFlag: true})
	in := []any{
		map[string]any{"id": "a1", "name": "Alice"},
		map[string]any{"id": "b1", "name": "Bob"},
	}
	if err := w.Render(in); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header+2 rows, got %d: %q", len(lines), out.String())
	}
	if !strings.HasPrefix(lines[0], "id\t") {
		t.Fatalf("expected `id` first column, got %q", lines[0])
	}
}

func TestRender_Human_KeyValueForBigObject(t *testing.T) {
	w, out, _ := newTestWriter(Options{NoColor: true})
	obj := map[string]any{
		"id": "psc_abc", "to": "alice", "from": "bob", "carrier": "USPS", "status": "rendered",
	}
	if err := w.Render(obj); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "id  ") || !strings.Contains(out.String(), "psc_abc") {
		t.Fatalf("kv out = %q", out.String())
	}
}

func TestSelect_FlatField(t *testing.T) {
	w, out, _ := newTestWriter(Options{JSONFlag: true, Select: "id"})
	if err := w.Render(map[string]any{"id": "x", "name": "y"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"id": "x"`) || strings.Contains(out.String(), `"name"`) {
		t.Fatalf("select did not project: %q", out.String())
	}
}

func TestSelect_DotPath(t *testing.T) {
	w, out, _ := newTestWriter(Options{JSONFlag: true, Select: "to.city"})
	in := map[string]any{
		"to": map[string]any{"city": "SF", "state": "CA"},
		"id": "x",
	}
	if err := w.Render(in); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, `"city": "SF"`) || strings.Contains(got, `"state"`) || strings.Contains(got, `"id"`) {
		t.Fatalf("dot-path projection wrong: %q", got)
	}
}

func TestNDJSON_StreamSlice(t *testing.T) {
	w, out, _ := newTestWriter(Options{})
	w.mode = NDJSON
	in := []any{
		map[string]any{"id": "1"},
		map[string]any{"id": "2"},
	}
	if err := w.Render(in); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 ndjson lines, got %d: %q", len(lines), out.String())
	}
}

func TestNotice_QuietSuppressed(t *testing.T) {
	w, _, errBuf := newTestWriter(Options{Quiet: true})
	w.Notice("hello %s", "world")
	if errBuf.Len() != 0 {
		t.Fatalf("quiet did not suppress: %q", errBuf.String())
	}
}

func TestNotice_DefaultEmits(t *testing.T) {
	w, _, errBuf := newTestWriter(Options{})
	w.Notice("hello")
	if !strings.Contains(errBuf.String(), "hello") {
		t.Fatalf("notice not emitted: %q", errBuf.String())
	}
}
