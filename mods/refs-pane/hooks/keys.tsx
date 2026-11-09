// refs-pane keys: the Client module that hears the arrow keys and Enter.
// - draws the one-line hint ("click here ...") a person clicks to give it the focus
// - onKey: left/right post {tab: -1|1}, up/down post {move: -1|1}, return posts {open: true}
// - holds no state and reads nothing; hooks/register.tsx handles the posts at `ui.message`

import type { ClientModule } from 'claude-code'

const HINT = 'click here, then ←/→ tabs · ↑/↓ rows · enter opens; esc releases'

const keys: ClientModule = (_props, surface) => {
  surface.onKey(e => {
    if (e.ctrl || e.meta) return
    if (e.key === 'left') surface.post({ tab: -1 })
    else if (e.key === 'right') surface.post({ tab: 1 })
    else if (e.key === 'up') surface.post({ move: -1 })
    else if (e.key === 'down') surface.post({ move: 1 })
    else if (e.key === 'return') surface.post({ open: true })
  })
  return surface.elements.Text({ dimColor: true, wrap: 'truncate-end', children: HINT })
}

export default keys
