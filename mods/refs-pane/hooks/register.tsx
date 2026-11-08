// refs-pane: a live pane of the session project's refs, opened by /refs-pane.
// - finds the store: <git toplevel of the session cwd, else the cwd>/.ai/refs.json
// - polls it on a clock (stat mtime+size) and redraws when it changes; reads it only on a change
// - draws a header (path, total, count per kind), refs grouped under kind headings (kinds order, `unknown` last),
//   one Button per ref, and a detail card for the selected one
// - keys: j/k move the cursor, o opens the selected ref's home, Enter on the selected row does the same
//   (Enter on another row selects it)
// - the one action opens the home file with the OS opener (open on darwin, xdg-open elsewhere); the pane never writes refs.json
// - a missing file draws an empty state; malformed JSON draws an error line

import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register, Timer } from 'claude-code'

import type { Entry, RefsFile, Snap } from '../types'

const PANE = 'refs-pane'
const POLL_MS = 2000
const UNKNOWN = 'unknown'

const INIT: Snap = { status: 'init', path: '', error: null, kindOrder: [], entries: [], sig: '' }
const snap = atom({ plugin: 'refs-pane', key: 'snap' } as const, INIT)
const selected = atom({ plugin: 'refs-pane', key: 'selected' } as const, null)
const note = atom({ plugin: 'refs-pane', key: 'note' } as const, null)
const root = atom({ plugin: 'refs-pane', key: 'root' } as const, null)

let timer: Timer | null = null

function firstLine(s: string): string {
  return s.split('\n')[0] ?? ''
}

// The project root: git toplevel of the session cwd, else the cwd itself.
async function findRoot($: EngineInterface): Promise<string> {
  const cwd = await $.session.cwd()
  try {
    const ran = await $.process.run(['git', 'rev-parse', '--show-toplevel'], { cwd })
    const top = ran.stdout.trim()
    if (ran.exitCode === 0 && top !== '') return top
  } catch {
    // no git: the cwd stands
  }
  return cwd
}

async function rootOf($: EngineInterface): Promise<string> {
  const known = await read($, root)
  if (known !== null) return known
  const found = await findRoot($)
  await update($, root, () => found)
  return found
}

// Kind names in order: the file's kinds (insertion order), then names only the refs carry, `unknown` last.
function kindOrderOf(file: RefsFile, entries: readonly Entry[]): string[] {
  const order: string[] = []
  const add = (name: string) => {
    if (name !== UNKNOWN && !order.includes(name)) order.push(name)
  }
  for (const name of Object.values(file.kinds ?? {})) add(name)
  for (const e of entries) add(e.kind)
  return order
}

function parse(text: string, path: string, sig: string): Snap {
  let file: RefsFile
  try {
    file = JSON.parse(text) as RefsFile
  } catch (err) {
    const why = err instanceof Error ? err.message : String(err)
    return { ...INIT, status: 'error', path, sig, error: `malformed JSON: ${firstLine(why)}` }
  }
  if (typeof file !== 'object' || file === null || Array.isArray(file)) {
    return { ...INIT, status: 'error', path, sig, error: 'malformed JSON: expected an object' }
  }
  const refs = file.refs ?? {}
  if (typeof refs !== 'object' || refs === null || Array.isArray(refs)) {
    return { ...INIT, status: 'error', path, sig, error: 'malformed JSON: "refs" is not an object' }
  }
  const entries: Entry[] = Object.entries(refs).map(([key, r]) => ({
    ...r,
    key,
    kind: typeof r?.kind === 'string' && r.kind !== '' ? r.kind : UNKNOWN,
    value: typeof r?.value === 'string' ? r.value : String(r?.value ?? ''),
  }))
  return { status: 'ok', path, error: null, kindOrder: kindOrderOf(file, entries), entries, sig }
}

// Stats the store and reads it only when mtime+size moved; redraws on a change.
async function poll($: EngineInterface): Promise<void> {
  const path = `${await rootOf($)}/.ai/refs.json`
  const before = await read($, snap)
  let next: Snap
  let sig: string
  try {
    if (!(await $.fs.exists(path))) {
      sig = 'missing'
      if (before.status === 'missing' && before.path === path) return
      next = { ...INIT, status: 'missing', path, sig }
    } else {
      const st = await $.fs.stat(path)
      sig = `${st.mtimeMs}:${st.size}`
      if (before.status !== 'init' && before.path === path && before.sig === sig) return
      next = parse(await $.fs.read(path), path, sig)
    }
  } catch (err) {
    const why = err instanceof Error ? err.message : String(err)
    sig = `error:${firstLine(why)}`
    if (before.status === 'error' && before.sig === sig) return
    next = { ...INIT, status: 'error', path, sig, error: `cannot read: ${firstLine(why)}` }
  }
  await update($, snap, () => next)
  const keys = next.entries.map(e => e.key)
  const top = groupsOf(next)[0]?.entries[0]?.key ?? null
  await update($, selected, cur => (cur !== null && keys.includes(cur) ? cur : top))
  try {
    $.ui.invalidate('ui.render')
  } catch {
    // no surface attached
  }
}

type Group = { kind: string; entries: Entry[] }

function groupsOf(s: Snap): Group[] {
  const names = [...s.kindOrder, UNKNOWN]
  return names
    .map(kind => ({ kind, entries: s.entries.filter(e => e.kind === kind) }))
    .filter(g => g.entries.length > 0)
}

function shownValue(e: Entry): string {
  return (e.title && e.url ? e.title : e.value).replace(/\s+/g, ' ')
}

function clip(s: string, room: number): string {
  if (room < 1) return ''
  return s.length <= room ? s : `${s.slice(0, Math.max(0, room - 1))}…`
}

// `path[:LINE]` -> the path; the line is dropped for the OS opener.
function homePath(home: string): string {
  return home.replace(/:\d+$/, '')
}

async function move($: EngineInterface, by: number): Promise<void> {
  const flat = groupsOf(await read($, snap)).flatMap(g => g.entries)
  if (flat.length === 0) return
  const sel = await read($, selected)
  const cur = flat.findIndex(e => e.key === sel)
  const at = Math.max(0, Math.min(flat.length - 1, (cur < 0 ? 0 : cur) + by))
  await update($, selected, () => flat[at]?.key ?? null)
  await update($, note, () => null)
}

async function select($: EngineInterface, key: string): Promise<void> {
  await update($, selected, () => key)
  await update($, note, () => null)
}

// The one action: open the selected ref's home with the OS opener.
async function openHome($: EngineInterface): Promise<void> {
  const s = await read($, snap)
  const key = await read($, selected)
  const entry = s.entries.find(e => e.key === key)
  if (entry === undefined) return
  if (!entry.home) {
    await update($, note, () => `${entry.key} has no home to open.`)
    return
  }
  const rel = homePath(entry.home)
  const abs = rel.startsWith('/') ? rel : `${await rootOf($)}/${rel}`
  try {
    const uname = await $.process.run(['uname', '-s'])
    const opener = uname.stdout.trim() === 'Darwin' ? 'open' : 'xdg-open'
    const ran = await $.process.run([opener, abs])
    await update($, note, () => (ran.exitCode === 0 ? `Opened ${abs}` : `${opener} failed: ${firstLine(ran.stderr)}`))
  } catch (err) {
    await update($, note, () => `could not open ${abs}: ${firstLine(err instanceof Error ? err.message : String(err))}`)
  }
}

function startTimer($: EngineInterface): void {
  timer?.cancel()
  timer = $.clock.every(POLL_MS, () => void poll($).catch(() => undefined))
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({
      name: 'refs-pane',
      description: "Show this project's refs in a live pane",
    })
    startTimer($)

    return next(e)
  })

  on('command.run', { command: 'refs-pane' }, async $ => {
    startTimer($)
    await update($, root, () => null)
    await poll($)
    await $.ui.open({ id: PANE, title: 'refs', focus: true })

    return { text: 'refs pane opened.' }
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Box, Text, Button } = $.ui.resolve(e)
    if ((await read($, snap)).status === 'init') await poll($)
    const s = await read($, snap)
    const key = await read($, selected)
    const msg = await read($, note)
    const width = Math.max(20, Math.min(200, e.props.bodyColumns))

    const head = (
      <Text bold wrap="truncate-end">
        {s.path === '' ? 'refs' : s.path}
      </Text>
    )

    if (s.status === 'error') {
      return (
        <Box flexDirection="column">
          {head}
          <Text color="red" wrap="wrap">
            {s.error ?? 'error'}
          </Text>
        </Box>
      )
    }
    if (s.status === 'missing' || (s.status === 'ok' && s.entries.length === 0)) {
      return (
        <Box flexDirection="column">
          {head}
          <Text dimColor wrap="wrap">
            no refs yet — run `refs seed`
          </Text>
        </Box>
      )
    }

    const groups = groupsOf(s)
    const counts = groups.map(g => `${g.kind} ${g.entries.length}`).join(' · ')
    const current = s.entries.find(r => r.key === key) ?? groups[0]?.entries[0]

    return (
      <Box flexDirection="column">
        {head}
        <Text wrap="truncate-end">
          {s.entries.length} refs · {counts}
        </Text>
        {groups.map(g => (
          <Box key={`kind:${g.kind}`} flexDirection="column">
            <Text bold dimColor>
              {g.kind} ({g.entries.length})
            </Text>
            {g.entries.map(r => {
              const isSel = current?.key === r.key
              const mark = r.home ? ' ⌂' : ''
              const lead = `${isSel ? '▸' : ' '} ${r.key}  `
              const label = `${lead}${clip(shownValue(r), width - lead.length - mark.length - 1)}${mark}`
              return (
                <Button
                  key={`ref:${r.key}`}
                  label={label}
                  plain
                  onPress={() => void (isSel ? openHome($) : select($, r.key))}
                />
              )
            })}
          </Box>
        ))}
        {current !== undefined && (
          <Box key="detail" flexDirection="column" borderStyle="round" paddingX={1}>
            <Text bold>
              {current.key} <Text dimColor>{current.kind}</Text>
            </Text>
            <Text wrap="wrap">
              {current.value}
            </Text>
            {current.title && <Text>title: {current.title}</Text>}
            {current.url && (
              <Text wrap="wrap">
                url: {current.url}
              </Text>
            )}
            <Text wrap="wrap">
              home: {current.home ?? '—'}
            </Text>
            <Text>set_at: {current.set_at ?? '—'}</Text>
            {msg !== null && (
              <Text color="yellow" wrap="wrap">
                {msg}
              </Text>
            )}
          </Box>
        )}
        <Text dimColor>
          j/k move · o or enter opens home
        </Text>
        <Box display="none" flexDirection="row">
          <Button key="next" label="down" hotkey="j" plain onPress={() => void move($, 1)} />
          <Button key="prev" label="up" hotkey="k" plain onPress={() => void move($, -1)} />
          <Button key="open" label="open home" hotkey="o" plain onPress={() => void openHome($)} />
        </Box>
      </Box>
    )
  })
}
