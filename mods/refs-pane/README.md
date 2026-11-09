# refs-pane

A live pane of the session project's refs, opened with `/refs-pane`.

Reads `<git toplevel of the session cwd, else cwd>/.ai/refs.json` (the store of the `refs` CLI) and redraws within
two seconds of a change. The pane never writes the file.

- Header: the file path, then a tab bar: `all`, then one tab per kind, each with its count.
- Body: the tab's refs grouped under kind headings (kinds order, `unknown` last). Each line is key, value (a link shows its
  title) and a home marker. Long values truncate to the pane width.
- Detail card for the selected ref: full value, title, url, home, set_at.
- Missing file: "no refs yet — run `refs seed`". Malformed JSON: one error line.

**Keys** (pane focused)

	h / l          previous / next tab (wraps)
	j / k          move the cursor within the tab
	o              open the selected ref's home
	Enter          on the selected row: open its home; on another row: select it
	Tab, arrows    walk the rows, scroll the pane

Opening uses `open` on macOS and `xdg-open` elsewhere, on the home path under the project root (a `:LINE` suffix is
dropped). A ref with no home shows a note instead.

**Install**

Listed in the refs marketplace (this repo). With the marketplace added:

	claude plugin install refs-pane@refs --scope user
