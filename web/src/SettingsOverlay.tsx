import { useEffect, useId, useRef, useState } from 'react'
import type { GameState, TermsInfo } from './api'
import {
  addTerms,
  fetchTerms,
  removeTerms,
  replaceTerms,
  resetAll,
  resetBoard,
  resetTerms,
  setName,
} from './api'
import { LOOKS, type Look, type Theme } from './theme'
import { parseTermsText } from './utils'

type Section = 'appearance' | 'terms' | 'player' | 'danger'

type SettingsOverlayProps = {
  open: boolean
  onClose: () => void
  state: GameState
  theme: Theme
  look: Look
  onThemeChange: (theme: Theme) => void
  onLookChange: (look: Look) => void
  onStateChange: (state: GameState) => void
  onTermsChanged: () => void
}

export function SettingsOverlay({
  open,
  onClose,
  state,
  theme,
  look,
  onThemeChange,
  onLookChange,
  onStateChange,
  onTermsChanged,
}: SettingsOverlayProps) {
  const titleId = useId()
  const panelRef = useRef<HTMLDivElement>(null)
  const [section, setSection] = useState<Section>('terms')
  const [termsInfo, setTermsInfo] = useState<TermsInfo | null>(null)
  const [termsError, setTermsError] = useState('')
  const [termsBusy, setTermsBusy] = useState(false)
  const [search, setSearch] = useState('')
  const [addText, setAddText] = useState('')
  const [importText, setImportText] = useState('')
  const [showImport, setShowImport] = useState(false)
  const [nameDraft, setNameDraft] = useState(state.name)
  const [nameError, setNameError] = useState('')
  const [nameBusy, setNameBusy] = useState(false)
  const [dangerError, setDangerError] = useState('')
  const [dangerBusy, setDangerBusy] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!open) return
    setNameDraft(state.name)
    setSection('appearance')
    setTermsError('')
    setNameError('')
    setDangerError('')
    setShowImport(false)
    setImportText('')
    setAddText('')
    setSearch('')
    void loadTerms()
    const prev = document.activeElement as HTMLElement | null
    panelRef.current?.focus()
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('keydown', onKey)
      prev?.focus?.()
    }
    // Intentionally only re-init when the overlay opens.
  }, [open])

  useEffect(() => {
    setNameDraft(state.name)
  }, [state.name])

  async function loadTerms() {
    setTermsBusy(true)
    setTermsError('')
    try {
      setTermsInfo(await fetchTerms())
    } catch (err) {
      setTermsError(err instanceof Error ? err.message : 'Failed to load terms')
    } finally {
      setTermsBusy(false)
    }
  }

  async function withTerms(action: () => Promise<TermsInfo>) {
    setTermsBusy(true)
    setTermsError('')
    try {
      const next = await action()
      setTermsInfo(next)
      onTermsChanged()
    } catch (err) {
      setTermsError(err instanceof Error ? err.message : 'Terms update failed')
    } finally {
      setTermsBusy(false)
    }
  }

  if (!open) return null

  const filtered =
    termsInfo?.terms.filter((t) => t.toLowerCase().includes(search.trim().toLowerCase())) ?? []

  return (
    <div className="settings-backdrop" onClick={onClose} role="presentation">
      <div
        ref={panelRef}
        className="settings-panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
      >
        <header className="settings-header">
          <h2 id={titleId}>Settings</h2>
          <div className="settings-header-actions">
            <button type="button" className="ghost-btn" onClick={onClose} aria-label="Close settings">
              Close
            </button>
          </div>
        </header>

        <nav className="settings-nav" aria-label="Settings sections">
          {(
            [
              ['appearance', 'Appearance'],
              ['terms', 'Terms'],
              ['player', 'Player'],
              ['danger', 'Danger'],
            ] as const
          ).map(([id, label]) => (
            <button
              key={id}
              type="button"
              className={section === id ? 'nav-tab active' : 'nav-tab'}
              onClick={() => setSection(id)}
            >
              {label}
            </button>
          ))}
        </nav>

        <div className="settings-body">
          {section === 'appearance' && (
            <section className="settings-section">
              <p className="section-lead">Pick a look and light or dark mode. Changes apply immediately.</p>

              <div>
                <p className="field-label" id="look-label">
                  Look
                </p>
                <div className="look-grid" role="radiogroup" aria-labelledby="look-label">
                  {LOOKS.map((option) => (
                    <button
                      key={option.id}
                      type="button"
                      role="radio"
                      aria-checked={look === option.id}
                      className={look === option.id ? 'look-card active' : 'look-card'}
                      onClick={() => onLookChange(option.id)}
                    >
                      <span className="look-card-title">{option.label}</span>
                      <span className="look-card-blurb">{option.blurb}</span>
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <p className="field-label" id="mode-label">
                  Mode
                </p>
                <div className="mode-toggle" role="radiogroup" aria-labelledby="mode-label">
                  {(
                    [
                      ['light', 'Light'],
                      ['dark', 'Dark'],
                    ] as const
                  ).map(([id, label]) => (
                    <button
                      key={id}
                      type="button"
                      role="radio"
                      aria-checked={theme === id}
                      className={theme === id ? 'mode-btn active' : 'mode-btn'}
                      onClick={() => onThemeChange(id)}
                    >
                      {label}
                    </button>
                  ))}
                </div>
              </div>
            </section>
          )}

          {section === 'terms' && (
            <section className="settings-section">
              <p className="section-lead">
                Manage the term pool used when reshuffling. Changes apply to the next board.
                {termsInfo && (
                  <>
                    {' '}
                    <span className="pill">{termsInfo.custom ? 'Custom file' : 'Defaults'}</span>
                    <span className="muted"> · {termsInfo.count} terms</span>
                  </>
                )}
              </p>

              {termsError && <p className="error-banner">{termsError}</p>}

              <div className="field-row">
                <input
                  type="search"
                  className="text-input"
                  placeholder="Filter terms…"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  disabled={!termsInfo}
                />
              </div>

              <ul className="terms-list" aria-label="Terms">
                {termsBusy && !termsInfo && <li className="muted">Loading…</li>}
                {filtered.map((term) => (
                  <li key={term}>
                    <span className="term-text">{term}</span>
                    <button
                      type="button"
                      className="text-btn danger-text"
                      disabled={termsBusy}
                      onClick={() => void withTerms(() => removeTerms([term]))}
                    >
                      Remove
                    </button>
                  </li>
                ))}
                {termsInfo && filtered.length === 0 && <li className="muted">No matching terms</li>}
              </ul>

              <form
                className="stack-form"
                onSubmit={(e) => {
                  e.preventDefault()
                  const parts = parseTermsText(addText)
                  if (parts.length === 0) return
                  void withTerms(async () => {
                    const next = await addTerms(parts)
                    setAddText('')
                    return next
                  })
                }}
              >
                <label className="field-label" htmlFor="add-terms">
                  Add terms
                </label>
                <textarea
                  id="add-terms"
                  className="text-area"
                  rows={3}
                  placeholder="One term per line"
                  value={addText}
                  onChange={(e) => setAddText(e.target.value)}
                  disabled={termsBusy}
                />
                <button type="submit" className="primary-btn" disabled={termsBusy || !addText.trim()}>
                  Add
                </button>
              </form>

              <div className="action-row">
                <button
                  type="button"
                  className="ghost-btn"
                  disabled={termsBusy}
                  onClick={() => setShowImport((v) => !v)}
                >
                  {showImport ? 'Hide import' : 'Import / replace'}
                </button>
                <button
                  type="button"
                  className="ghost-btn"
                  disabled={termsBusy || !termsInfo}
                  onClick={() => {
                    if (!termsInfo) return
                    const body = `# Custom bingo terms\n${termsInfo.terms.join('\n')}\n`
                    const blob = new Blob([body], { type: 'text/plain' })
                    const url = URL.createObjectURL(blob)
                    const a = document.createElement('a')
                    a.href = url
                    a.download = 'terms.txt'
                    a.click()
                    URL.revokeObjectURL(url)
                  }}
                >
                  Download
                </button>
                <button
                  type="button"
                  className="ghost-btn"
                  disabled={termsBusy || !termsInfo?.custom}
                  onClick={() => {
                    if (!window.confirm('Reset to embedded default terms?')) return
                    void withTerms(() => resetTerms())
                  }}
                >
                  Reset to defaults
                </button>
              </div>

              {showImport && (
                <div className="import-box">
                  <p className="muted small">
                    Paste a full term list (at least 24 lines) or choose a <code>.txt</code> file. This
                    replaces your custom terms file.
                  </p>
                  <textarea
                    className="text-area"
                    rows={6}
                    placeholder="# comments ok&#10;Term one&#10;Term two"
                    value={importText}
                    onChange={(e) => setImportText(e.target.value)}
                    disabled={termsBusy}
                  />
                  <div className="action-row">
                    <input
                      ref={fileRef}
                      type="file"
                      accept=".txt,text/plain"
                      hidden
                      onChange={(e) => {
                        const file = e.target.files?.[0]
                        if (!file) return
                        void file.text().then((text) => setImportText(text))
                        e.target.value = ''
                      }}
                    />
                    <button
                      type="button"
                      className="ghost-btn"
                      disabled={termsBusy}
                      onClick={() => fileRef.current?.click()}
                    >
                      Choose file
                    </button>
                    <button
                      type="button"
                      className="primary-btn"
                      disabled={termsBusy || parseTermsText(importText).length < 24}
                      onClick={() =>
                        void withTerms(async () => {
                          const next = await replaceTerms(parseTermsText(importText))
                          setShowImport(false)
                          setImportText('')
                          return next
                        })
                      }
                    >
                      Replace all
                    </button>
                  </div>
                </div>
              )}
            </section>
          )}

          {section === 'player' && (
            <section className="settings-section">
              <p className="section-lead">Display name shown above the board.</p>
              {nameError && <p className="error-banner">{nameError}</p>}
              <form
                className="stack-form"
                onSubmit={(e) => {
                  e.preventDefault()
                  setNameBusy(true)
                  setNameError('')
                  void setName(nameDraft)
                    .then((next) => {
                      onStateChange(next)
                    })
                    .catch((err) => {
                      setNameError(err instanceof Error ? err.message : 'Could not save name')
                    })
                    .finally(() => setNameBusy(false))
                }}
              >
                <label className="field-label" htmlFor="player-name">
                  Name
                </label>
                <input
                  id="player-name"
                  className="text-input"
                  maxLength={64}
                  value={nameDraft}
                  onChange={(e) => setNameDraft(e.target.value)}
                  disabled={nameBusy}
                  required
                />
                <button type="submit" className="primary-btn" disabled={nameBusy || !nameDraft.trim()}>
                  Save name
                </button>
              </form>
            </section>
          )}

          {section === 'danger' && (
            <section className="settings-section">
              <p className="section-lead">Reshuffle today&apos;s board or wipe all local bingo data.</p>
              {dangerError && <p className="error-banner">{dangerError}</p>}
              <div className="danger-actions">
                <button
                  type="button"
                  className="ghost-btn"
                  disabled={dangerBusy}
                  onClick={() => {
                    setDangerBusy(true)
                    setDangerError('')
                    void resetBoard()
                      .then((next) => {
                        onStateChange(next)
                        onClose()
                      })
                      .catch((err) => {
                        setDangerError(err instanceof Error ? err.message : 'Reset failed')
                      })
                      .finally(() => setDangerBusy(false))
                  }}
                >
                  Reshuffle board
                </button>
                <button
                  type="button"
                  className="danger-btn"
                  disabled={dangerBusy}
                  onClick={() => {
                    const ok = window.confirm(
                      'This deletes your name, custom terms, session, and wins. Continue?',
                    )
                    if (!ok) return
                    setDangerBusy(true)
                    setDangerError('')
                    void resetAll()
                      .then((next) => {
                        onStateChange(next)
                        void loadTerms()
                        onClose()
                      })
                      .catch((err) => {
                        setDangerError(err instanceof Error ? err.message : 'Reset all failed')
                      })
                      .finally(() => setDangerBusy(false))
                  }}
                >
                  Reset everything
                </button>
              </div>
            </section>
          )}
        </div>
      </div>
    </div>
  )
}
