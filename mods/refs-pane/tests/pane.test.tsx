// pane.test: refs-pane against a mocked file system and process.
// - refs group under kind headings in kinds order, `unknown` last; links show their title
// - the tab bar counts all and each kind; h/l switch tabs and wrap; a tab shows only its kind
// - the selected ref's detail shows value, title/url, home, set_at; j/k move the cursor
// - a ref with no home gets a note on open
// - open-home calls the OS opener with the absolute path (no :LINE)
// - a missing file draws the empty state; malformed JSON draws an error line
// - the pane redraws when refs.json changes
// - arrows (through the Client in hooks/keys.tsx): left/right switch tab, up/down move the cursor, return opens the home;
//   ctrl-arrows and other keys do nothing; the hint line is drawn

import { expect, mock, test } from 'claude-code/testing'
import type { On } from 'claude-code'
import type { Engine } from 'claude-code/testing'

const PANE = {
  plugin: 'refs-pane',
  component: 'Pane',
  requestId: 'refs-pane',
  props: {
    title: 'refs',
    isFocused: true,
    bodyColumns: 40,
    placement: 'dock',
    scroll: { offset: 0, bodyRows: 30 },
    view: {},
  },
} as const

const RUN = {
  command: 'refs-pane',
  args: '',
  origin: { kind: 'composer' },
  presentation: { isFullscreen: false, columns: 120 },
} as const

const PATH = '/repo/.ai/refs.json'

const FILE = {
  kinds: { R: 'requirement', 'pr#': 'pr' },
  refs: {
    'pr#1': {
      kind: 'pr',
      value: '[Add the thing](https://example.com/pr/1)',
      title: 'Add the thing',
      url: 'https://example.com/pr/1',
      set_at: '2026-11-01T10:00:00',
    },
    R1: { kind: 'requirement', value: 'Never lose a ref', home: 'docs/canon/REQUIREMENTS.md:12', set_at: '2026-11-02T10:00:00' },
    X9: { kind: 'unknown', value: 'stray' },
    R2: { kind: 'requirement', value: 'x'.repeat(200) },
  },
}

const ok = (stdout: string) => ({ value: { exitCode: 0, stdout, stderr: '', isStdoutTruncated: false, isStderrTruncated: false } })

// The engine beneath the pane: a file at PATH (or none), a git root of /repo, a recording process.
function world(on: On, files: { text: string | null }, calls: string[][]) {
  const clock = mock.clock(on)
  on('session.cwd', () => ({ value: '/repo/sub' }))
  on('process.run', (_$, e) => {
    calls.push([...e.argv])
    if (e.argv[0] === 'git') return ok('/repo\n')
    if (e.argv[0] === 'uname') return ok('Darwin\n')
    return ok('')
  })
  on('fs.exists', () => ({ value: files.text !== null }))
  on('fs.stat', () => ({ value: { kind: 'file', size: (files.text ?? '').length, mtimeMs: 1, isLink: false } }))
  on('fs.read', () => ({ value: files.text ?? '' }))
  on('command.register', () => ({ value: { command: 'refs-pane' } }))
  on('ui.open', () => ({ value: { isPlaced: true } }))
  return clock
}

async function mounted($: Engine, on: On, text: string | null, calls: string[][] = [], surface: 'terminal' | 'desktop' = 'terminal') {
  const files = { text }
  const clock = world(on, files, calls)
  await $.command.run(RUN)
  const ui = await $.ui.mount({ ...PANE, surface })
  return { ui, files, clock }
}

for (const surface of ['terminal', 'desktop'] as const) {
  test(`refs group by kind in kinds order, unknown last (${surface})`, async ($, on) => {
    const { ui } = await mounted($, on, JSON.stringify(FILE), [], surface)

    const headings = ['requirement (2)', 'pr (1)', 'unknown (1)']
    const texts = (await ui.findAll({ type: 'Text' })).map(t => t.text)
    const at = headings.map(h => texts.indexOf(h))
    expect(at[0]).toBeGreaterThanOrEqual(0)
    expect(at[0]).toBeLessThan(at[1] as number)
    expect(at[1]).toBeLessThan(at[2] as number)
    // Rows follow group order, insertion order within a group.
    const rows = (await ui.findAll({ type: 'Button' })).map(b => b.key)
    expect(rows.filter(k => k?.startsWith('ref:'))).toEqual(['ref:R1', 'ref:R2', 'ref:pr#1', 'ref:X9'])
    // A link shows its title, not the markdown.
    expect((await ui.find({ key: 'ref:pr#1' }))?.text).toContain('Add the thing')
    expect((await ui.find({ key: 'ref:pr#1' }))?.text).not.toContain('https://')
    await ui.unmount()
  })
}

test('the tab bar counts all and each kind, All first and selected', async ($, on) => {
  const { ui } = await mounted($, on, JSON.stringify(FILE))

  expect((await ui.find({ type: 'Text', text: /\/repo\/\.ai\/refs\.json/ }))?.text).toBe(PATH)
  expect((await ui.find({ key: 'tab:all' }))?.text).toContain('[all 4]')
  expect((await ui.find({ key: 'tab:requirement' }))?.text).toContain('requirement 2')
  expect((await ui.find({ key: 'tab:pr' }))?.text).toContain('pr 1')
  expect((await ui.find({ key: 'tab:unknown' }))?.text).toContain('unknown 1')
  await ui.unmount()
})

test('h and l switch tabs and wrap; a tab shows only its kind; a click on a tab shows it', async ($, on) => {
  const { ui } = await mounted($, on, JSON.stringify(FILE))

  await ui.press({ key: 'tabnext' })
  expect((await ui.find({ key: 'tab:requirement' }))?.text).toContain('[requirement 2]')
  expect(await ui.find({ key: 'ref:pr#1' })).toBeUndefined()
  expect(await ui.find({ key: 'ref:R1' })).toBeDefined()
  await ui.press({ key: 'next' })
  await ui.press({ key: 'next' })
  expect((await ui.find({ type: 'Text', text: /^(R1|R2|pr#1|X9) / }))?.text).toContain('R2')

  await ui.press({ key: 'tabprev' })
  await ui.press({ key: 'tabprev' })
  expect((await ui.find({ key: 'tab:unknown' }))?.text).toContain('[unknown 1]')
  expect((await ui.find({ type: 'Text', text: /^(R1|R2|pr#1|X9) / }))?.text).toContain('X9')

  await ui.press({ key: 'tab:pr' })
  expect((await ui.find({ key: 'tab:pr' }))?.text).toContain('[pr 1]')
  expect((await ui.find({ type: 'Text', text: /^(R1|R2|pr#1|X9) / }))?.text).toContain('pr#1')
  await ui.unmount()
})

test('a long value truncates to the pane width, and a home shows a marker', async ($, on) => {
  const { ui } = await mounted($, on, JSON.stringify(FILE))

  const long = (await ui.find({ key: 'ref:R2' }))?.text ?? ''
  expect(long.length).toBeLessThanOrEqual(40)
  expect(long).toContain('…')
  expect((await ui.find({ key: 'ref:R1' }))?.text).toContain('⌂')
  await ui.unmount()
})

test('the selected ref shows its detail; j and k move the cursor', async ($, on) => {
  const { ui } = await mounted($, on, JSON.stringify(FILE))

  // First in group order is R1.
  expect((await ui.find({ type: 'Text', text: /^(R1|R2|pr#1|X9) / }))?.text).toContain('R1')
  expect((await ui.find({ type: 'Text', text: /^Never lose a ref$/ }))?.text).toBe('Never lose a ref')
  expect((await ui.find({ type: 'Text', text: /^home: / }))?.text).toContain('docs/canon/REQUIREMENTS.md:12')
  expect((await ui.find({ type: 'Text', text: /^set_at: / }))?.text).toContain('2026-11-02T10:00:00')

  await ui.press({ key: 'next' })
  await ui.press({ key: 'next' })
  expect((await ui.find({ type: 'Text', text: /^(R1|R2|pr#1|X9) / }))?.text).toContain('pr#1')
  expect((await ui.find({ type: 'Text', text: /^title: / }))?.text).toContain('Add the thing')
  expect((await ui.find({ type: 'Text', text: /^url: / }))?.text).toContain('https://example.com/pr/1')

  await ui.press({ key: 'prev' })
  expect((await ui.find({ type: 'Text', text: /^(R1|R2|pr#1|X9) / }))?.text).toContain('R2')
  await ui.unmount()
})

test('open on a ref with no home shows a note and spawns no opener', async ($, on) => {
  const calls: string[][] = []
  const { ui } = await mounted($, on, JSON.stringify(FILE), calls)

  await ui.press({ key: 'next' })
  await ui.press({ key: 'open' })
  expect((await ui.find({ type: 'Text', text: /has no home|Opened|failed/ }))?.text).toMatch(/R2 has no home/)
  expect(calls.some(c => c[0] === 'open' || c[0] === 'xdg-open')).toBe(false)
  await ui.unmount()
})

test('open-home runs the OS opener on the absolute path, without the line', async ($, on) => {
  const calls: string[][] = []
  const { ui } = await mounted($, on, JSON.stringify(FILE), calls)

  await ui.press({ key: 'open' })
  expect(calls).toContainEqual(['open', '/repo/docs/canon/REQUIREMENTS.md'])
  await ui.unmount()
})

test('Enter on the selected row opens it; on another row it selects', async ($, on) => {
  const calls: string[][] = []
  const { ui } = await mounted($, on, JSON.stringify(FILE), calls)

  await ui.press({ key: 'ref:R2' })
  expect((await ui.find({ type: 'Text', text: /^(R1|R2|pr#1|X9) / }))?.text).toContain('R2')
  expect(calls.some(c => c[0] === 'open')).toBe(false)
  await ui.press({ key: 'ref:R2' })
  expect((await ui.find({ type: 'Text', text: /has no home|Opened|failed/ }))?.text).toMatch(/no home/)
  await ui.unmount()
})

test('a missing file draws the empty state', async ($, on) => {
  const { ui } = await mounted($, on, null)

  expect((await ui.find({ type: 'Text', text: /no refs yet/ }))?.text).toBe('no refs yet — run `refs seed`')
  await ui.unmount()
})

test('malformed JSON draws an error line, not a crash', async ($, on) => {
  const { ui } = await mounted($, on, '{"refs": ')

  expect((await ui.find({ type: 'Text', text: /^malformed JSON/ }))?.text).toMatch(/^malformed JSON/)
  await ui.unmount()
})

test('the pane redraws when refs.json changes', async ($, on) => {
  const { ui, files, clock } = await mounted($, on, JSON.stringify({ kinds: {}, refs: { A1: { kind: 'unknown', value: 'one' } } }))
  expect((await ui.find({ key: 'tab:all' }))?.text).toContain('all 1')

  files.text = JSON.stringify({ kinds: {}, refs: { A1: { kind: 'unknown', value: 'one' }, A2: { kind: 'unknown', value: 'two' } } })
  await clock.advance(2000)

  expect((await ui.find({ key: 'tab:all' }))?.text).toContain('all 2')
  expect(await ui.find({ key: 'ref:A2' })).toBeDefined()
  await ui.unmount()
})

const ROW = /^(R1|R2|pr#1|X9) /

test('left and right arrows switch tabs and wrap', async ($, on) => {
  const { ui } = await mounted($, on, JSON.stringify(FILE))

  await ui.key({ key: 'right' })
  expect((await ui.find({ key: 'tab:requirement' }))?.text).toContain('[requirement 2]')
  await ui.key({ key: 'left' })
  await ui.key({ key: 'left' })
  expect((await ui.find({ key: 'tab:unknown' }))?.text).toContain('[unknown 1]')
  await ui.unmount()
})

test('up and down arrows move the cursor and stop at the ends', async ($, on) => {
  const { ui } = await mounted($, on, JSON.stringify(FILE))

  await ui.key({ key: 'down' })
  await ui.key({ key: 'down' })
  expect((await ui.find({ type: 'Text', text: ROW }))?.text).toContain('pr#1')
  await ui.key({ key: 'up' })
  expect((await ui.find({ type: 'Text', text: ROW }))?.text).toContain('R2')
  await ui.key({ key: 'up' })
  await ui.key({ key: 'up' })
  expect((await ui.find({ type: 'Text', text: ROW }))?.text).toContain('R1')
  await ui.unmount()
})

test('return opens the selected ref home; other keys do nothing', async ($, on) => {
  const calls: string[][] = []
  const { ui } = await mounted($, on, JSON.stringify(FILE), calls)

  await ui.key({ key: 'x' })
  await ui.key({ key: 'right', ctrl: true })
  expect(calls.some(c => c[0] === 'open')).toBe(false)
  expect((await ui.find({ key: 'tab:all' }))?.text).toContain('[all 4]')

  await ui.key({ key: 'return' })
  expect(calls).toContainEqual(['open', '/repo/docs/canon/REQUIREMENTS.md'])

  await ui.key({ key: 'down' })
  await ui.key({ key: 'return' })
  expect((await ui.find({ type: 'Text', text: /has no home/ }))?.text).toMatch(/R2 has no home/)
  await ui.unmount()
})

test('the hint line tells the person to click for arrows; hotkeys still work beside it', async ($, on) => {
  const { ui } = await mounted($, on, JSON.stringify(FILE))

  expect((await ui.find({ type: 'Text', text: /click here/, in: 'keys' }))?.text).toMatch(/←\/→ tabs/)
  await ui.press({ key: 'tabnext' })
  expect((await ui.find({ key: 'tab:requirement' }))?.text).toContain('[requirement 2]')
  await ui.unmount()
})
