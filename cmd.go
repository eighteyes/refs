// cmd.go — the refs command: list (optionally filtered by --kind), show, set, rm, kind, kinds, home, ids.
//
// Responsibilities:
//   - parse flags by hand (--kind, --home, --clear, with `=` forms)
//   - dispatch the first positional to a subcommand, else treat it as a KEY
//   - validate keys (RESERVED, '-', leading digit) and parse [title](url) values
//   - print results to stdout, errors and nags to stderr
package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var reserved = map[string]bool{"rm": true, "kind": true, "kinds": true, "home": true, "seed": true, "hook": true, "ids": true, "sync": true, "guard": true}

var linkRE = regexp.MustCompile(`^\[([^\]]+)\]\(([^)\s\p{Z}\x{85}\x{1c}-\x{1f}]+)\)\n?$`)

func now() string { return time.Now().Format("2006-01-02T15:04:05-0700") }

// refText is the one-line rendering of a ref's value; links read as `title — url`.
func refText(r *omap) string {
	if u := r.Get("url"); truthy(u) {
		t := r.Get("title")
		if !truthy(t) {
			t = u
		}
		return pystr(t) + " — " + pystr(u)
	}
	if v, ok := r.Lookup("value"); ok {
		return pystr(v)
	}
	return ""
}

func (a *app) out(format string, args ...any) { fmt.Fprintf(a.stdout, format, args...) }

func (a *app) cmdRefs(argv []string) int {
	var kind string
	var home string
	hasHome, clear := false, false
	var pos []string
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		switch {
		case arg == "--kind":
			if i+1 >= len(argv) {
				a.die("--kind needs a NAME")
			}
			kind = argv[i+1]
			i++
		case strings.HasPrefix(arg, "--kind="):
			kind = arg[len("--kind="):]
		case arg == "--home":
			if i+1 >= len(argv) {
				a.die("--home needs a PATH")
			}
			home, hasHome = argv[i+1], true
			i++
		case strings.HasPrefix(arg, "--home="):
			home, hasHome = arg[len("--home="):], true
		case arg == "--clear":
			clear = true
		default:
			pos = append(pos, arg)
		}
	}

	root := projectRoot("")
	if err := autoSync(root, a.stderr); err != nil {
		fmt.Fprintln(a.stderr, "refs: canon auto-sync failed: "+err.Error())
	}
	store := loadRefs(root)
	kinds, refs := asObj(store.Get("kinds")), asObj(store.Get("refs"))
	save := func() {
		if err := saveRefs(root, store); err != nil {
			a.die("cannot write " + refsPath(root) + ": " + err.Error())
		}
	}
	noRef := func(key string) { a.die("no ref '" + key + "' in " + refsPath(root)) }

	if len(pos) == 0 {
		if refs.Len() == 0 {
			a.out("no refs in %s\n", refsPath(root))
			return 0
		}
		// --kind NAME with no positionals filters by kind name or key prefix
		keys := refs.keys
		if kind != "" {
			keys = nil
			for _, k := range refs.keys {
				kv, _ := asObj(refs.Get(k)).Lookup("kind")
				if pystr(kv) == kind || prefixOf(k) == kind {
					keys = append(keys, k)
				}
			}
			if len(keys) == 0 {
				fmt.Fprintf(a.stderr, "refs: no refs of kind or prefix '%s'\n", kind)
				return 0
			}
		}
		w, kw := 0, 0
		for _, k := range keys {
			w = max(w, utf8.RuneCountInString(k))
			kv, ok := asObj(refs.Get(k)).Lookup("kind")
			if !ok {
				kv = ""
			}
			kw = max(kw, utf8.RuneCountInString(pystr(kv)))
		}
		for _, key := range keys {
			r := asObj(refs.Get(key))
			mark := ""
			if h := r.Get("home"); truthy(h) {
				mark = "  @ " + pystr(h)
			}
			k := "unknown"
			if kv, ok := r.Lookup("kind"); ok {
				k = pystr(kv)
			}
			a.out("%s  %s  %s%s\n", padRight(k, kw), padRight(key, w), refText(r), mark)
		}
		return 0
	}

	head, args := pos[0], pos[1:]

	if clear && head != "home" {
		a.die("--clear applies only to: refs home KEY --clear")
	}

	switch head {
	case "kinds":
		if kinds.Len() == 0 {
			a.out("no kinds defined\n")
		}
		for _, p := range kinds.keys {
			a.out("%s  %s\n", p, pystr(kinds.Get(p)))
		}
		return 0

	case "kind":
		if len(args) != 2 {
			a.die("usage: refs kind PREFIX NAME")
		}
		prefix, name := args[0], args[1]
		kinds.Set(prefix, name)
		for _, key := range refs.keys {
			r, ok := refs.Get(key).(*omap)
			if ok && r.Get("kind") == "unknown" && prefixOf(key) == prefix {
				r.Set("kind", name)
			}
		}
		save()
		a.out("kind %s = %s\n", prefix, name)
		return 0

	case "ids":
		if len(args) == 0 {
			a.out("%s\n", idsMode(store))
			return 0
		}
		if len(args) != 1 || (args[0] != "fixed" && args[0] != "fluid") {
			a.die("usage: refs ids [fixed|fluid]")
		}
		store.Set("ids", args[0])
		save()
		a.out("ids = %s\n", args[0])
		return 0

	case "home":
		if clear {
			if len(args) != 1 {
				a.die("usage: refs home KEY --clear")
			}
		} else if len(args) != 2 {
			a.die("usage: refs home KEY PATH | refs home KEY --clear")
		}
		key := args[0]
		if _, ok := refs.Lookup(key); !ok {
			noRef(key)
		}
		r := asObj(refs.Get(key))
		if clear {
			r.Delete("home")
			save()
			a.out("home %s cleared\n", key)
			return 0
		}
		h := a.normalizeHome(root, args[1])
		r.Set("home", h)
		save()
		a.out("home %s = %s\n", key, h)
		return 0

	case "rm":
		if len(args) != 1 {
			a.die("usage: refs rm KEY")
		}
		key := args[0]
		if _, ok := refs.Lookup(key); !ok {
			noRef(key)
		}
		refs.Delete(key)
		save()
		a.out("rm %s\n", key)
		return 0
	}

	key := head
	if len(args) == 0 {
		r, ok := refs.Get(key).(*omap)
		if !ok || r.Len() == 0 {
			noRef(key)
		}
		k := "unknown"
		if kv, ok := r.Lookup("kind"); ok {
			k = pystr(kv)
		}
		a.out("%s  [%s]\n", key, k)
		v := ""
		if vv, ok := r.Lookup("value"); ok {
			v = pystr(vv)
		}
		a.out("  value   %s\n", v)
		if truthy(r.Get("url")) {
			a.out("  title   %s\n", pystr(r.Get("title")))
			a.out("  url     %s\n", pystr(r.Get("url")))
		}
		if h := r.Get("home"); truthy(h) {
			hs := pystr(h)
			hpath, _ := homeParts(hs)
			full := hpath
			if !filepath.IsAbs(hpath) {
				full = filepath.Join(root, hpath)
			}
			a.out("  home    %s\n", hs)
			a.out("  open    file://%s\n", full)
		}
		a.out("  set_at  %s\n", pystr(r.Get("set_at")))
		return 0
	}

	// set / overwrite
	if reserved[key] {
		a.die("'" + key + "' is a refs subcommand and cannot be a key")
	}
	if strings.HasPrefix(key, "-") {
		a.die("key '" + key + "' looks like a flag; keys cannot start with '-'")
	}
	if first, _ := utf8.DecodeRuneInString(key); key == "" || unicode.IsDigit(first) {
		a.die("key '" + key + "' must start with a non-digit prefix (e.g. R20, pr#131, adr-3)")
	}
	value := strings.Join(args, " ")
	prefix := prefixOf(key)
	storedHome := ""
	if hasHome {
		storedHome = a.normalizeHome(root, home)
	}
	if kind != "" {
		kinds.Set(prefix, kind)
	}
	var resolved any = "unknown"
	if kv, ok := kinds.Lookup(prefix); ok {
		resolved = kv
	}
	entry := newOmap()
	entry.Set("kind", resolved)
	entry.Set("value", value)
	entry.Set("set_at", now())
	if m := linkRE.FindStringSubmatch(value); m != nil {
		entry.Set("title", m[1])
		entry.Set("url", m[2])
	}
	var keep any = storedHome
	if storedHome == "" {
		keep = asObj(refs.Get(key)).Get("home")
	}
	if truthy(keep) {
		entry.Set("home", keep)
	}
	refs.Set(key, entry)
	save()
	a.out("set %s [%s]\n", key, pystr(resolved))
	if resolved == "unknown" {
		fmt.Fprintf(a.stderr, "refs: prefix '%s' has no kind. Define it: refs kind %s NAME\n", prefix, prefix)
	}
	return 0
}
