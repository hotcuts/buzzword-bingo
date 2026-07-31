import { useEffect, useState } from 'react'
import type { GameState } from './api'
import { fetchState, markCell, resetBoard } from './api'
import { BingoBoard } from './BingoBoard'
import { OnboardingOverlay } from './OnboardingOverlay'
import { SettingsOverlay } from './SettingsOverlay'
import { sfx } from './sfx'
import {
  applyAppearance,
  getPreferredLook,
  getPreferredTheme,
  type Look,
  type Theme,
} from './theme'
import { formatPeriodLabel, hasSelections } from './utils'

export default function App() {
  const [state, setState] = useState<GameState | null>(null)
  const [error, setError] = useState('')
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [theme, setTheme] = useState<Theme>(() => getPreferredTheme())
  const [look, setLook] = useState<Look>(() => getPreferredLook())
  const [termsHint, setTermsHint] = useState(false)

  useEffect(() => {
    applyAppearance(look, theme)
  }, [look, theme])

  useEffect(() => {
    void fetchState()
      .then((next) => {
        setState(next)
      })
      .catch(() => {
        setError('Could not load bingo board. Is Bingo still running?')
      })
  }, [])

  async function handleMark(index: number) {
    if (!state) return
    await sfx.unlock()
    const wasMarked = state.marked[index]
    const wasWon = state.won
    try {
      const next = await markCell(index)
      if (next.marked[index] && !wasMarked) await sfx.mark()
      else if (!next.marked[index] && wasMarked) await sfx.unmark()
      if (next.won && !wasWon) await sfx.win()
      setState(next)
    } catch (err) {
      console.error(err)
    }
  }

  async function doReset(confirmIfMarked: boolean) {
    if (!state) return
    if (confirmIfMarked && hasSelections(state.cells, state.marked)) {
      const ok = window.confirm(
        'You have selections on this board. Resetting will reshuffle and clear them. Continue?',
      )
      if (!ok) return
    }
    await sfx.unlock()
    try {
      const next = await resetBoard()
      await sfx.unmark()
      setState(next)
      setTermsHint(false)
    } catch (err) {
      console.error(err)
    }
  }

  const needsOnboarding = Boolean(state && !state.name.trim())

  return (
    <div className="app-shell">
      <main>
        <header className="top-bar">
          <div className="brand-block">
            <h1>Buzzword Bingo</h1>
            {state?.name ? <p className="player">{state.name}</p> : null}
            {state && !needsOnboarding && (
              <div className="meta">
                <div className="meta-item">
                  <span className="meta-label">{state.period === 'weekly' ? 'Week' : 'Date'}</span>
                  <span className="meta-value">{formatPeriodLabel(state.date, state.period)}</span>
                </div>
                <div className="meta-item">
                  <span className="meta-label">Wins</span>
                  <span className="meta-value">{state.winCount}</span>
                </div>
              </div>
            )}
          </div>
          {!needsOnboarding && (
            <button
              type="button"
              className="menu-btn"
              aria-label="Open settings"
              onClick={() => setSettingsOpen(true)}
            >
              <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false" className="menu-icon">
                <path
                  fill="currentColor"
                  d="M19.14 12.94c.04-.31.06-.63.06-.94s-.02-.63-.06-.94l2.03-1.58a.49.49 0 0 0 .12-.61l-1.92-3.32a.49.49 0 0 0-.59-.22l-2.39.96a7.07 7.07 0 0 0-1.63-.94l-.36-2.54a.48.48 0 0 0-.48-.41h-3.84a.48.48 0 0 0-.48.41l-.36 2.54c-.59.24-1.13.55-1.63.94l-2.39-.96a.49.49 0 0 0-.59.22L2.77 8.87a.48.48 0 0 0 .12.61l2.03 1.58c-.04.31-.06.63-.06.94s.02.63.06.94L2.89 14.52a.49.49 0 0 0-.12.61l1.92 3.32c.13.22.39.3.59.22l2.39-.96c.5.39 1.04.71 1.63.94l.36 2.54c.05.23.25.41.48.41h3.84c.23 0 .43-.18.48-.41l.36-2.54c.59-.24 1.13-.55 1.63-.94l2.39.96c.22.08.46 0 .59-.22l1.92-3.32a.48.48 0 0 0-.12-.61l-2.03-1.58zM12 15.6A3.6 3.6 0 1 1 12 8.4a3.6 3.6 0 0 1 0 7.2z"
                />
              </svg>
            </button>
          )}
        </header>

        {termsHint && (
          <div className="terms-hint" role="status">
            <span>Term pool updated. Reshuffle to use the new terms on the board.</span>
            <button type="button" className="text-btn" onClick={() => void doReset(true)}>
              Reshuffle now
            </button>
            <button type="button" className="text-btn" onClick={() => setTermsHint(false)}>
              Dismiss
            </button>
          </div>
        )}

        {error && <p className="error-banner load-error">{error}</p>}

        {state && !needsOnboarding && (
          <div className="stage">
            <BingoBoard
              cells={state.cells}
              marked={state.marked}
              won={state.won}
              onMark={(i) => void handleMark(i)}
              onNav={() => void sfx.nav()}
              onPlayAgain={() => void doReset(false)}
            />
            <aside className="side-actions">
              <button
                type="button"
                className="side-reset"
                title="Reshuffle board"
                aria-label="Reshuffle board"
                onClick={() => void doReset(true)}
              >
                <svg className="reset-icon" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                  <path
                    fill="currentColor"
                    d="M17.65 6.35A7.95 7.95 0 0 0 12 4V1L7 6l5 5V7c2.76 0 5 2.24 5 5s-2.24 5-5 5-5-2.24-5-5H4c0 4.42 3.58 8 8 8s8-3.58 8-8c0-2.21-.9-4.21-2.35-5.65z"
                  />
                </svg>
              </button>
            </aside>
          </div>
        )}
      </main>

      {needsOnboarding && (
        <OnboardingOverlay
          theme={theme}
          look={look}
          onThemeChange={setTheme}
          onLookChange={setLook}
          onComplete={(next) => {
            setState(next)
            setTermsHint(false)
            setSettingsOpen(false)
          }}
        />
      )}

      {state && !needsOnboarding && (
        <SettingsOverlay
          open={settingsOpen}
          onClose={() => setSettingsOpen(false)}
          state={state}
          theme={theme}
          look={look}
          onThemeChange={setTheme}
          onLookChange={setLook}
          onStateChange={setState}
          onTermsChanged={() => setTermsHint(true)}
        />
      )}
    </div>
  )
}
