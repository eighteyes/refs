# refs-pane

A live pane of the session project's refs, opened with `/refs-pane`.

Reads `<git toplevel of the session cwd, else cwd>/.ai/refs.json` (the store of the `refs` CLI) and redraws within
two seconds of a change. The pane never writes the file.

- Header: the file path, the total, the count per kind.
- Body: refs grouped under kind headings (kinds order, `unknown` last). Each line is key, value (a link shows its
  title) and a home marker. Long values truncate to the pane width.
- Detail card for the selected ref: full value, title, url, home, set_at.
- Missing file: "no refs yet — run `refs seed`". Malformed JSON: one error line.

**Keys** (pane focused)

	j / k          move the cursor
	o              open the selected ref's home
	Enter          on the selected row: open its home; on another row: select it
	Tab, arrows    walk the rows, scroll the pane

Opening uses `open` on macOS and `xdg-open` elsewhere, on the home path under the project root (a `:LINE` suffix is
dropped). A ref with no home shows a note instead.
