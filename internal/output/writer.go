package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
)

// Writer is the single entry point for emitting command results in any mode.
// Construct one per command from the resolved Options.
type Writer struct {
	mode        Mode
	stdout      io.Writer
	stderr      io.Writer
	color       bool
	quiet       bool
	resultsOnly bool
	selectPaths []string
}

// New constructs a Writer with the resolved mode and selection paths.
func New(opts Options) *Writer {
	mode := opts.Resolve()
	w := &Writer{
		mode:        mode,
		stdout:      opts.Stdout,
		stderr:      opts.Stderr,
		quiet:       opts.Quiet,
		resultsOnly: opts.ResultsOnly,
		color:       mode == Human && !opts.NoColor && isTerminal(opts.Stdout),
	}
	if opts.Select != "" {
		for _, p := range strings.Split(opts.Select, ",") {
			if p = strings.TrimSpace(p); p != "" {
				w.selectPaths = append(w.selectPaths, p)
			}
		}
	}
	return w
}

// Mode returns the active output mode.
func (w *Writer) Mode() Mode { return w.mode }

// Stderr returns the writer for human-only side channels (progress, hints).
func (w *Writer) Stderr() io.Writer { return w.stderr }

// Notice prints a human-readable message to stderr. It is suppressed in --quiet.
func (w *Writer) Notice(format string, args ...any) {
	if w.quiet {
		return
	}
	_, _ = fmt.Fprintf(w.stderr, format+"\n", args...)
}

// Render emits a result in the active mode. v should be a struct, map, or
// slice — anything json.Marshal accepts. Pointers are dereferenced.
func (w *Writer) Render(v any) error {
	if v == nil {
		return nil
	}
	v = w.project(v)

	switch w.mode {
	case JSON:
		return w.renderJSON(v)
	case Plain:
		return w.renderPlain(v)
	case NDJSON:
		return w.renderNDJSON(v)
	default:
		return w.renderHuman(v)
	}
}

// RenderID prints just the resource ID — useful for `--quiet` mode or when the
// caller only wants the bare identifier to pipe into another command.
func (w *Writer) RenderID(id string) error {
	_, err := fmt.Fprintln(w.stdout, id)
	if err != nil {
		return fmt.Errorf("write id: %w", err)
	}
	return nil
}

// project applies --select dot-path projection to v. Returns v unchanged when
// no paths are configured.
func (w *Writer) project(v any) any {
	if len(w.selectPaths) == 0 {
		return v
	}
	buf, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var generic any
	if err := json.Unmarshal(buf, &generic); err != nil {
		return v
	}
	return selectFields(generic, w.selectPaths)
}

func (w *Writer) renderJSON(v any) error {
	enc := json.NewEncoder(w.stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func (w *Writer) renderNDJSON(v any) error {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return w.renderJSONLine(v)
	}
	for i := 0; i < rv.Len(); i++ {
		if err := w.renderJSONLine(rv.Index(i).Interface()); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) renderJSONLine(v any) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal ndjson: %w", err)
	}
	if _, err := fmt.Fprintln(w.stdout, string(buf)); err != nil {
		return fmt.Errorf("write ndjson: %w", err)
	}
	return nil
}

func (w *Writer) renderPlain(v any) error {
	rows, headers := flatten(v)
	if len(rows) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w.stdout, strings.Join(headers, "\t")); err != nil {
		return fmt.Errorf("write plain header: %w", err)
	}
	for _, row := range rows {
		cells := make([]string, len(headers))
		for i, h := range headers {
			cells[i] = row[h]
		}
		if _, err := fmt.Fprintln(w.stdout, strings.Join(cells, "\t")); err != nil {
			return fmt.Errorf("write plain row: %w", err)
		}
	}
	return nil
}

func (w *Writer) renderHuman(v any) error {
	rows, headers := flatten(v)
	switch {
	case len(rows) == 0:
		return nil
	case len(rows) == 1 && len(headers) > 3:
		return w.renderKeyValue(rows[0], headers)
	default:
		return w.renderTable(rows, headers)
	}
}

func (w *Writer) renderKeyValue(row map[string]string, headers []string) error {
	maxKey := 0
	for _, k := range headers {
		if len(k) > maxKey {
			maxKey = len(k)
		}
	}
	for _, k := range headers {
		line := fmt.Sprintf("%-*s  %s\n", maxKey, k, row[k])
		if _, err := fmt.Fprint(w.stdout, line); err != nil {
			return fmt.Errorf("write kv: %w", err)
		}
	}
	return nil
}

func (w *Writer) renderTable(rows []map[string]string, headers []string) error {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, r := range rows {
		for i, h := range headers {
			if l := len(r[h]); l > widths[i] {
				widths[i] = l
			}
		}
	}
	if err := writeRow(w.stdout, headers, widths); err != nil {
		return err
	}
	sep := make([]string, len(headers))
	for i, wd := range widths {
		sep[i] = strings.Repeat("-", wd)
	}
	if err := writeRow(w.stdout, sep, widths); err != nil {
		return err
	}
	for _, r := range rows {
		cells := make([]string, len(headers))
		for i, h := range headers {
			cells[i] = r[h]
		}
		if err := writeRow(w.stdout, cells, widths); err != nil {
			return err
		}
	}
	return nil
}

func writeRow(out io.Writer, cells []string, widths []int) error {
	for i, c := range cells {
		sep := "  "
		if i == len(cells)-1 {
			sep = "\n"
		}
		if _, err := fmt.Fprintf(out, "%-*s%s", widths[i], c, sep); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}
	return nil
}

// flatten turns any JSON-marshalable v into a list of string-string rows with a
// stable header order. Single objects yield one row. Slices yield one row per
// element, with the union of keys as headers (sorted, but `id` first if present).
func flatten(v any) (rows []map[string]string, headers []string) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, nil
	}
	var generic any
	if err := json.Unmarshal(buf, &generic); err != nil {
		return nil, nil
	}
	switch x := generic.(type) {
	case []any:
		for _, item := range x {
			row := flattenOne(item)
			if row == nil {
				continue
			}
			rows = append(rows, row)
		}
	case map[string]any:
		row := flattenOne(x)
		if row != nil {
			rows = append(rows, row)
		}
	default:
		rows = append(rows, map[string]string{"value": fmt.Sprint(x)})
	}
	if len(rows) == 0 {
		return nil, nil
	}
	keys := map[string]struct{}{}
	for _, r := range rows {
		for k := range r {
			keys[k] = struct{}{}
		}
	}
	headers = make([]string, 0, len(keys))
	for k := range keys {
		headers = append(headers, k)
	}
	sort.Slice(headers, func(i, j int) bool {
		ai, aj := headers[i], headers[j]
		if ai == "id" {
			return true
		}
		if aj == "id" {
			return false
		}
		return ai < aj
	})
	return rows, headers
}

func flattenOne(v any) map[string]string {
	m, ok := v.(map[string]any)
	if !ok {
		return map[string]string{"value": fmt.Sprint(v)}
	}
	out := make(map[string]string, len(m))
	for k, val := range m {
		out[k] = stringify(val)
	}
	return out
}

func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool, float64, float32, int, int32, int64:
		return fmt.Sprint(x)
	default:
		buf, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprint(x)
		}
		return string(buf)
	}
}

// ErrPipeBroken is reported when downstream consumer (e.g. head) closes the
// pipe. Commands should treat it as success.
var ErrPipeBroken = errors.New("pipe closed by reader")
