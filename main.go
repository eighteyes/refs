// main.go — entry point and shared plumbing for the refs CLI.
//
// Responsibilities:
//   - dispatch argv: --help, `hook`, `seed`, everything else to the refs command
//   - own the app type (stdin/stdout/stderr) so tests can drive the CLI in-process
//   - die(): print `refs: ...` to stderr and unwind with an exit code
//   - hold the HELP text, written for a model reader
package main

import (
	"io"
	"os"
)

const help = `refs — per-project key-value store for the pillars of your session
(requirements, PRs, decisions). Stored in <project>/.ai/refs.json.

WHEN
  Run ` + "`refs`" + ` at session start and after compaction to recall the pillars.
  Add one when a requirement, PR, or decision starts to bound your work.
  ` + "`refs rm KEY`" + ` when it is done. Read on demand; nothing is injected.
  Run ` + "`refs seed`" + ` when the project has no refs, for how to populate them.

USAGE
  refs                              list refs
  refs --kind NAME|PREFIX           list only refs of that kind or key prefix
                                    (refs --kind truth, refs --kind TRU-)
  refs KEY                          show one
  refs KEY VALUE... [--kind NAME] [--home PATH[:LINE]]   set / overwrite
  refs rm KEY
  refs kind PREFIX NAME             name a key prefix; ` + "`refs kinds`" + ` lists them
  refs home KEY PATH[:LINE]         set the source-of-truth file
  refs home KEY --clear
  refs ids [fixed|fluid]            show or set whether canon IDs may change
  refs seed                         print how to populate an empty project
  refs hook                         SessionStart hook (stdin payload, JSON out)

KEYS
  Key prefix is the kind: R20 -> R, pr#131 -> pr#. Keys start with a non-digit
  and not '-'. Reserved: rm kind kinds home seed hook ids.
  Unknown prefix saves as kind=unknown and nags; define it with --kind or
  ` + "`refs kind PREFIX NAME`" + `.
  VALUE of the form [title](url) is parsed into title and url.

IDS
  Canon IDs (TRU-03, DEC-14...) are fluid or fixed, per project. Stored as
  "ids" in refs.json; absent means fluid. ` + "`refs ids`" + ` prints the mode.
  fluid: the tighten skill may renumber, delete, reuse and reorder canon IDs.
  fixed: IDs never change and are never reused.
  Projects start fluid; run ` + "`refs ids fixed`" + ` once the canon settles.

HOME
  --home PATH[:LINE] names the file that is the source of truth for a ref.
  Stored root-relative; must live inside the project. Read the home before
  trusting the value. Setting a value without --home keeps the existing home.
`

type app struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// exit unwinds the stack from die() back to run().
type exit struct{ code int }

func (a *app) die(msg string) {
	io.WriteString(a.stderr, "refs: "+msg+"\n")
	panic(exit{1})
}

// run executes the CLI and returns the process exit code.
func run(argv []string, stdin io.Reader, stdout, stderr io.Writer) (code int) {
	a := &app{stdin: stdin, stdout: stdout, stderr: stderr}
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(exit); ok {
				code = e.code
				return
			}
			panic(r)
		}
	}()
	if len(argv) > 0 {
		switch argv[0] {
		case "-h", "--help", "help":
			io.WriteString(stdout, help)
			return 0
		case "hook":
			return a.cmdHook()
		case "seed":
			if len(argv) > 1 {
				a.die("usage: refs seed")
			}
			io.WriteString(stdout, hookEmpty+"\n")
			return 0
		}
	}
	return a.cmdRefs(argv)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
