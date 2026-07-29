export type Theme = 'light' | 'dark'
export type Look = 'meeting' | 'neon' | 'editorial'

export const LOOKS: readonly {
  id: Look
  label: string
  blurb: string
}[] = [
  { id: 'meeting', label: 'Meeting', blurb: 'Clean neutrals for daily standup bingo.' },
  { id: 'neon', label: 'Neon', blurb: 'High contrast with punchy mark energy.' },
  { id: 'editorial', label: 'Editorial', blurb: 'Brand-forward type with atmospheric depth.' },
] as const

const THEME_KEY = 'bingo-theme'
const LOOK_KEY = 'bingo-look'

export function getPreferredTheme(): Theme {
  const stored = localStorage.getItem(THEME_KEY)
  if (stored === 'light' || stored === 'dark') return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function getPreferredLook(): Look {
  const stored = localStorage.getItem(LOOK_KEY)
  if (stored === 'meeting' || stored === 'neon' || stored === 'editorial') return stored
  return 'meeting'
}

export function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme
  localStorage.setItem(THEME_KEY, theme)
}

export function applyLook(look: Look) {
  document.documentElement.dataset.look = look
  localStorage.setItem(LOOK_KEY, look)
}

export function applyAppearance(look: Look, theme: Theme) {
  applyLook(look)
  applyTheme(theme)
}
