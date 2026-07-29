import { useId, useState } from 'react'
import type { GameState } from './api'
import { replaceTerms, resetBoard, setName } from './api'
import { LOOKS, type Look, type Theme } from './theme'
import { parseTermsText } from './utils'

const MIN_TERMS = 24

type Step = 'appearance' | 'name' | 'buzzwords'
type TermsChoice = 'defaults' | 'custom'

type OnboardingOverlayProps = {
  theme: Theme
  look: Look
  onThemeChange: (theme: Theme) => void
  onLookChange: (look: Look) => void
  onComplete: (state: GameState) => void
}

const STEPS: readonly { id: Step; label: string }[] = [
  { id: 'appearance', label: 'Look' },
  { id: 'name', label: 'Name' },
  { id: 'buzzwords', label: 'Words' },
]

export function OnboardingOverlay({
  theme,
  look,
  onThemeChange,
  onLookChange,
  onComplete,
}: OnboardingOverlayProps) {
  const titleId = useId()
  const [step, setStep] = useState<Step>('appearance')
  const [nameDraft, setNameDraft] = useState('')
  const [termsChoice, setTermsChoice] = useState<TermsChoice>('defaults')
  const [pasteText, setPasteText] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const parsedCount = parseTermsText(pasteText).length
  const stepIndex = STEPS.findIndex((s) => s.id === step)

  function goNext() {
    setError('')
    if (step === 'appearance') {
      setStep('name')
      return
    }
    if (step === 'name') {
      if (!nameDraft.trim()) {
        setError('Enter a display name to continue.')
        return
      }
      setStep('buzzwords')
      return
    }
  }

  function goBack() {
    setError('')
    if (step === 'name') setStep('appearance')
    else if (step === 'buzzwords') setStep('name')
  }

  async function finish() {
    setError('')
    const name = nameDraft.trim()
    if (!name) {
      setError('Enter a display name to continue.')
      setStep('name')
      return
    }

    if (termsChoice === 'custom') {
      const parsed = parseTermsText(pasteText)
      if (parsed.length < MIN_TERMS) {
        setError(`Need at least ${MIN_TERMS} buzzwords (you have ${parsed.length}). Custom list was not saved.`)
        return
      }
    }

    setBusy(true)
    try {
      let next = await setName(name)
      if (termsChoice === 'custom') {
        const parsed = parseTermsText(pasteText)
        await replaceTerms(parsed)
        next = await resetBoard()
      }
      onComplete(next)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not finish setup')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="onboard-backdrop" role="presentation">
      <div
        className="onboard-panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
      >
        <header className="onboard-header">
          <p className="onboard-eyebrow">Welcome</p>
          <h2 id={titleId}>Set up Buzzword Bingo</h2>
          <nav className="onboard-steps" aria-label="Setup steps">
            {STEPS.map((s, i) => (
              <span
                key={s.id}
                className={
                  i === stepIndex ? 'onboard-step active' : i < stepIndex ? 'onboard-step done' : 'onboard-step'
                }
              >
                {s.label}
              </span>
            ))}
          </nav>
        </header>

        <div className="onboard-body">
          {error && <p className="error-banner">{error}</p>}

          {step === 'appearance' && (
            <section className="settings-section">
              <p className="section-lead">Pick a look and light or dark mode. Changes apply immediately.</p>

              <div>
                <p className="field-label" id="onboard-look-label">
                  Look
                </p>
                <div className="look-grid" role="radiogroup" aria-labelledby="onboard-look-label">
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
                <p className="field-label" id="onboard-mode-label">
                  Mode
                </p>
                <div className="mode-toggle" role="radiogroup" aria-labelledby="onboard-mode-label">
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

          {step === 'name' && (
            <section className="settings-section">
              <p className="section-lead">This name shows above your board.</p>
              <div className="stack-form">
                <label className="field-label" htmlFor="onboard-name">
                  Display name
                </label>
                <input
                  id="onboard-name"
                  className="text-input"
                  maxLength={64}
                  value={nameDraft}
                  onChange={(e) => setNameDraft(e.target.value)}
                  autoFocus
                  required
                  disabled={busy}
                />
              </div>
            </section>
          )}

          {step === 'buzzwords' && (
            <section className="settings-section">
              <p className="section-lead">
                Use the built-in list, or paste your own (at least {MIN_TERMS}). You can add more later in
                Settings.
              </p>

              <div className="look-grid" role="radiogroup" aria-label="Buzzword source">
                <button
                  type="button"
                  role="radio"
                  aria-checked={termsChoice === 'defaults'}
                  className={termsChoice === 'defaults' ? 'look-card active' : 'look-card'}
                  onClick={() => {
                    setTermsChoice('defaults')
                    setError('')
                  }}
                  disabled={busy}
                >
                  <span className="look-card-title">Use defaults</span>
                  <span className="look-card-blurb">Ship with the embedded buzzword pool. No custom file.</span>
                </button>
                <button
                  type="button"
                  role="radio"
                  aria-checked={termsChoice === 'custom'}
                  className={termsChoice === 'custom' ? 'look-card active' : 'look-card'}
                  onClick={() => setTermsChoice('custom')}
                  disabled={busy}
                >
                  <span className="look-card-title">Paste my list</span>
                  <span className="look-card-blurb">
                    One term per line. Needs {MIN_TERMS}+ unique lines to save.
                  </span>
                </button>
              </div>

              {termsChoice === 'custom' && (
                <div className="stack-form">
                  <label className="field-label" htmlFor="onboard-terms">
                    Buzzwords · {parsedCount} / {MIN_TERMS} minimum
                  </label>
                  <textarea
                    id="onboard-terms"
                    className="text-area onboard-terms-area"
                    rows={10}
                    placeholder={'# comments ok\nSynergy\nCircle back\n…'}
                    value={pasteText}
                    onChange={(e) => {
                      setPasteText(e.target.value)
                      setError('')
                    }}
                    disabled={busy}
                  />
                </div>
              )}
            </section>
          )}
        </div>

        <footer className="onboard-footer">
          {step !== 'appearance' ? (
            <button type="button" className="ghost-btn" onClick={goBack} disabled={busy}>
              Back
            </button>
          ) : (
            <span />
          )}
          {step !== 'buzzwords' ? (
            <button type="button" className="primary-btn" onClick={goNext} disabled={busy}>
              Continue
            </button>
          ) : (
            <button
              type="button"
              className="primary-btn"
              onClick={() => void finish()}
              disabled={
                busy || (termsChoice === 'custom' && parsedCount < MIN_TERMS)
              }
            >
              {busy ? 'Saving…' : 'Start playing'}
            </button>
          )}
        </footer>
      </div>
    </div>
  )
}
