// refs-pane types: the refs.json shapes and the pane's kept state.
// - Ref, RefsFile: the on-disk shape the refs CLI writes
// - Snap: what the pane keeps of the last read (status, parsed refs, change signature)
// - declares the plugin's atoms on PluginState

export type Ref = {
  kind: string
  value: string
  set_at?: string
  title?: string
  url?: string
  // Root-relative `path[:LINE]`.
  home?: string
}

export type RefsFile = {
  kinds?: Record<string, string>
  refs?: Record<string, Ref>
}

export type Entry = Ref & { key: string }

// status: 'init' before the first read; 'ok'; 'missing' (no file); 'error' (unreadable or malformed JSON).
export type Snap = {
  status: 'init' | 'ok' | 'missing' | 'error'
  path: string
  error: string | null
  // Kind names in kinds-insertion order, as the file lists them.
  kindOrder: string[]
  entries: Entry[]
  // The stat signature `mtimeMs:size` the entries were read at.
  sig: string
}

declare module 'claude-code' {
  interface PluginState {
    'refs-pane': {
      snap: Snap
      selected: string | null
      note: string | null
      // The project root the store was found under; null until resolved.
      root: string | null
      // The tab shown: 'all' or a kind name.
      tab: string
    }
  }
}
