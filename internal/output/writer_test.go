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

func TestQuiet_EmitsBareIDsForObject(t *testing.T) {
	w, out, _ := newTestWriter(Options{Quiet: true})
	if err := w.Render(map[string]any{"id": "psc_abc", "status": "rendered"}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "psc_abc" {
		t.Fatalf("quiet object: got %q, want psc_abc", got)
	}
}

func TestQuiet_EmitsOnePerLineForList(t *testing.T) {
	w, out, _ := newTestWriter(Options{Quiet: true})
	in := []any{
		map[string]any{"id": "a1"},
		map[string]any{"id": "b1"},
	}
	if err := w.Render(in); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 2 || lines[0] != "a1" || lines[1] != "b1" {
		t.Fatalf("quiet list: got %q", out.String())
	}
}

func TestQuiet_FlattensDataEnvelope(t *testing.T) {
	w, out, _ := newTestWriter(Options{Quiet: true})
	in := map[string]any{"object": "list", "data": []any{
		map[string]any{"id": "x"},
		map[string]any{"id": "y"},
	}}
	if err := w.Render(in); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(out.String()); got != "x\ny" {
		t.Fatalf("quiet envelope: got %q", out.String())
	}
}

func TestResultsOnly_StripsEnvelope(t *testing.T) {
	w, out, _ := newTestWriter(Options{JSONFlag: true, ResultsOnly: true})
	in := map[string]any{
		"object": "list",
		"count":  2,
		"data":   []any{map[string]any{"id": "a"}, map[string]any{"id": "b"}},
	}
	if err := w.Render(in); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if strings.Contains(got, `"object"`) || strings.Contains(got, `"count"`) {
		t.Fatalf("results-only kept envelope: %q", got)
	}
	if !strings.Contains(got, `"id": "a"`) || !strings.Contains(got, `"id": "b"`) {
		t.Fatalf("results-only dropped data: %q", got)
	}
}

func TestResultsOnly_PassthroughForNonEnvelope(t *testing.T) {
	w, out, _ := newTestWriter(Options{JSONFlag: true, ResultsOnly: true})
	in := map[string]any{"id": "psc_only", "status": "rendered"}
	if err := w.Render(in); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"id": "psc_only"`) {
		t.Fatalf("results-only mangled non-envelope: %q", out.String())
	}
}
