// store.go — the refs.json store: ordered JSON, project root, home paths.
//
// Responsibilities:
//   - omap: an insertion-ordered JSON object (Go maps lose order)
//   - parse JSON into omap/[]any/string/json.Number/bool/nil, keeping key order
//   - serialize in Python json.dumps(indent=2) style, ASCII-escaped, so files stay compatible
//   - load (malformed or missing reads as empty) and atomically save <root>/.ai/refs.json
//   - find the project root: git toplevel, else cwd
//   - normalize `home` values: cwd-relative in, root-relative out, :LINE kept
//   - python-flavoured helpers (truthy, str) so odd hand-edited values behave as before
package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// omap is a JSON object that remembers key insertion order.
type omap struct {
	keys []string
	vals map[string]any
}

func newOmap() *omap { return &omap{vals: map[string]any{}} }

func (m *omap) Lookup(k string) (any, bool) { v, ok := m.vals[k]; return v, ok }
func (m *omap) Get(k string) any            { return m.vals[k] }
func (m *omap) Len() int                    { return len(m.keys) }

// Set replaces in place, or appends when the key is new.
func (m *omap) Set(k string, v any) {
	if _, ok := m.vals[k]; !ok {
		m.keys = append(m.keys, k)
	}
	m.vals[k] = v
}

func (m *omap) Delete(k string) {
	if _, ok := m.vals[k]; !ok {
		return
	}
	delete(m.vals, k)
	for i, kk := range m.keys {
		if kk == k {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func asObj(v any) *omap {
	if m, ok := v.(*omap); ok {
		return m
	}
	return newOmap()
}

// ---- parsing ----

func parseValue(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	d, isDelim := tok.(json.Delim)
	if !isDelim {
		return tok, nil // string, json.Number, bool, nil
	}
	switch d {
	case '{':
		m := newOmap()
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return nil, err
			}
			k, ok := kt.(string)
			if !ok {
				return nil, errors.New("object key is not a string")
			}
			v, err := parseValue(dec)
			if err != nil {
				return nil, err
			}
			m.Set(k, v)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return m, nil
	case '[':
		arr := []any{}
		for dec.More() {
			v, err := parseValue(dec)
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return arr, nil
	}
	return nil, errors.New("unexpected delimiter")
}

// parseJSON decodes one complete JSON document, preserving object key order.
func parseJSON(r io.Reader) (any, error) {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	v, err := parseValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("trailing data")
	}
	return v, nil
}

// ---- serializing ----

// quote renders a JSON string. ascii=true escapes non-ASCII like Python's default.
func quote(s string, ascii bool) string {
	const hex = "0123456789abcdef"
	var b strings.Builder
	u := func(c rune) {
		b.WriteString(`\u`)
		b.WriteByte(hex[c>>12&0xf])
		b.WriteByte(hex[c>>8&0xf])
		b.WriteByte(hex[c>>4&0xf])
		b.WriteByte(hex[c&0xf])
	}
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\b':
			b.WriteString(`\b`)
		case r == '\f':
			b.WriteString(`\f`)
		case r < 0x20:
			u(r)
		case r < 0x80 || !ascii:
			b.WriteRune(r)
		case r > 0xFFFF:
			r1, r2 := utf16.EncodeRune(r)
			u(r1)
			u(r2)
		default:
			u(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func writeValue(b *strings.Builder, v any, depth int) {
	pad := func(n int) { b.WriteString(strings.Repeat("  ", n)) }
	switch t := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if t {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		b.WriteString(quote(t, true))
	case json.Number:
		b.WriteString(t.String())
	case *omap:
		if t.Len() == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{\n")
		for i, k := range t.keys {
			pad(depth + 1)
			b.WriteString(quote(k, true))
			b.WriteString(": ")
			writeValue(b, t.vals[k], depth+1)
			if i < len(t.keys)-1 {
				b.WriteByte(',')
			}
			b.WriteByte('\n')
		}
		pad(depth)
		b.WriteByte('}')
	case []any:
		if len(t) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteString("[\n")
		for i, e := range t {
			pad(depth + 1)
			writeValue(b, e, depth+1)
			if i < len(t)-1 {
				b.WriteByte(',')
			}
			b.WriteByte('\n')
		}
		pad(depth)
		b.WriteByte(']')
	}
}

func serialize(v any) string {
	var b strings.Builder
	writeValue(&b, v, 0)
	b.WriteByte('\n')
	return b.String()
}

// ---- store ----

func refsPath(root string) string { return filepath.Join(root, ".ai", "refs.json") }

// loadRefs reads the store; missing or malformed files read as empty.
func loadRefs(root string) *omap {
	var store *omap
	if f, err := os.Open(refsPath(root)); err == nil {
		v, perr := parseJSON(f)
		f.Close()
		if perr == nil {
			store, _ = v.(*omap)
		}
	}
	if store == nil {
		store = newOmap()
	}
	for _, k := range []string{"kinds", "refs"} {
		if _, ok := store.Get(k).(*omap); !ok {
			store.Set(k, newOmap())
		}
	}
	return store
}

// saveRefs writes tmp then renames, so readers never see a partial file.
func saveRefs(root string, store *omap) error {
	path := refsPath(root)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(serialize(store)), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// ---- paths ----

// getcwd returns the physical working directory, like Python's os.getcwd().
func getcwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	if r, err := filepath.EvalSymlinks(cwd); err == nil {
		return r
	}
	return cwd
}

// projectRoot is the git toplevel of cwd, else cwd itself. Empty cwd means the process cwd.
func projectRoot(cwd string) string {
	if cwd == "" {
		cwd = getcwd()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if top := strings.TrimSpace(string(out)); err == nil && top != "" {
		return top
	}
	return cwd
}

// realpath resolves symlinks in the longest existing prefix of p.
func realpath(p string) string {
	p = filepath.Clean(p)
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	dir, base := filepath.Split(p)
	dir = filepath.Clean(dir)
	if dir == p || base == "" {
		return p
	}
	return filepath.Join(realpath(dir), base)
}

var homeLineRE = regexp.MustCompile(`^(.*):(\d+)$`)

// homeParts splits a home into (path, line-suffix); suffix is "" or ":N".
func homeParts(home string) (string, string) {
	if m := homeLineRE.FindStringSubmatch(home); m != nil {
		return m[1], ":" + m[2]
	}
	return home, ""
}

func expandUser(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			return h + p[1:]
		}
	}
	return p
}

// normalizeHome resolves raw (cwd-relative, optional :LINE) to a root-relative stored home.
func (a *app) normalizeHome(root, raw string) string {
	path, line := homeParts(raw)
	if path == "" {
		a.die("home needs a PATH")
	}
	path = expandUser(path)
	if !filepath.IsAbs(path) {
		path = filepath.Join(getcwd(), path)
	}
	abs := realpath(path)
	rel, err := filepath.Rel(realpath(root), abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		a.die("home '" + raw + "' is outside the project root " + root + "; homes must live inside the project")
	}
	if _, err := os.Stat(abs); err != nil {
		io.WriteString(a.stderr, "refs: home '"+raw+"' does not exist\n")
	}
	return rel + line
}

// ---- python-flavoured helpers ----

func truthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case json.Number:
		f, err := t.Float64()
		return err != nil || f != 0 || math.IsNaN(f)
	case *omap:
		return t.Len() > 0
	case []any:
		return len(t) > 0
	}
	return true
}

// pystr renders a value the way Python's str() would for scalars.
func pystr(v any) string {
	switch t := v.(type) {
	case nil:
		return "None"
	case bool:
		if t {
			return "True"
		}
		return "False"
	case string:
		return t
	case json.Number:
		return t.String()
	}
	return strings.TrimSpace(serialize(v))
}

// prefixOf returns the leading non-digit characters of key.
func prefixOf(key string) string {
	for i, r := range key {
		if unicode.IsDigit(r) {
			return key[:i]
		}
	}
	return key
}

func padRight(s string, w int) string {
	if n := utf8.RuneCountInString(s); n < w {
		return s + strings.Repeat(" ", w-n)
	}
	return s
}

func itoa(n int) string { return strconv.Itoa(n) }
