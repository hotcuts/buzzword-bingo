type AudioCtx = AudioContext

function noiseBuffer(audio: AudioCtx, seconds: number) {
  const n = Math.max(1, Math.floor(audio.sampleRate * seconds))
  const buf = audio.createBuffer(1, n, audio.sampleRate)
  const data = buf.getChannelData(0)
  for (let i = 0; i < n; i++) data[i] = Math.random() * 2 - 1
  return buf
}

function markSound(audio: AudioCtx) {
  const t0 = audio.currentTime

  const noise = audio.createBufferSource()
  noise.buffer = noiseBuffer(audio, 0.05)
  const bp = audio.createBiquadFilter()
  bp.type = 'bandpass'
  bp.frequency.setValueAtTime(900, t0)
  bp.Q.value = 4
  const ng = audio.createGain()
  ng.gain.setValueAtTime(0, t0)
  ng.gain.linearRampToValueAtTime(0.18, t0 + 0.005)
  ng.gain.exponentialRampToValueAtTime(0.001, t0 + 0.05)
  noise.connect(bp)
  bp.connect(ng)
  ng.connect(audio.destination)
  noise.start(t0)
  noise.stop(t0 + 0.05)

  const osc = audio.createOscillator()
  const g = audio.createGain()
  osc.type = 'sine'
  osc.frequency.setValueAtTime(420, t0)
  osc.frequency.exponentialRampToValueAtTime(880, t0 + 0.08)
  g.gain.setValueAtTime(0, t0)
  g.gain.linearRampToValueAtTime(0.13, t0 + 0.01)
  g.gain.exponentialRampToValueAtTime(0.001, t0 + 0.12)
  osc.connect(g)
  g.connect(audio.destination)
  osc.start(t0)
  osc.stop(t0 + 0.13)
}

function unmarkSound(audio: AudioCtx) {
  const t0 = audio.currentTime
  const osc = audio.createOscillator()
  const g = audio.createGain()
  osc.type = 'triangle'
  osc.frequency.setValueAtTime(520, t0)
  osc.frequency.exponentialRampToValueAtTime(180, t0 + 0.14)
  g.gain.setValueAtTime(0, t0)
  g.gain.linearRampToValueAtTime(0.12, t0 + 0.012)
  g.gain.exponentialRampToValueAtTime(0.001, t0 + 0.16)
  osc.connect(g)
  g.connect(audio.destination)
  osc.start(t0)
  osc.stop(t0 + 0.17)
}

function navSound(audio: AudioCtx) {
  const t0 = audio.currentTime
  const osc = audio.createOscillator()
  const g = audio.createGain()
  osc.type = 'sine'
  osc.frequency.setValueAtTime(760 + Math.random() * 80, t0)
  g.gain.setValueAtTime(0, t0)
  g.gain.linearRampToValueAtTime(0.045, t0 + 0.004)
  g.gain.exponentialRampToValueAtTime(0.001, t0 + 0.04)
  osc.connect(g)
  g.connect(audio.destination)
  osc.start(t0)
  osc.stop(t0 + 0.05)
}

function winSound(audio: AudioCtx) {
  const t0 = audio.currentTime
  const notes = [523.25, 659.25, 783.99, 1046.5]
  notes.forEach((freq, i) => {
    const start = t0 + i * 0.09
    const osc = audio.createOscillator()
    const g = audio.createGain()
    osc.type = i === notes.length - 1 ? 'triangle' : 'sine'
    osc.frequency.setValueAtTime(freq * 0.92, start)
    osc.frequency.exponentialRampToValueAtTime(freq, start + 0.04)
    g.gain.setValueAtTime(0, start)
    g.gain.linearRampToValueAtTime(0.12, start + 0.015)
    g.gain.exponentialRampToValueAtTime(0.001, start + 0.2)
    osc.connect(g)
    g.connect(audio.destination)
    osc.start(start)
    osc.stop(start + 0.22)
  })

  const shimmerAt = t0 + 0.28
  const sh = audio.createOscillator()
  const sg = audio.createGain()
  sh.type = 'sine'
  sh.frequency.setValueAtTime(1568, shimmerAt)
  sh.frequency.exponentialRampToValueAtTime(2093, shimmerAt + 0.12)
  sg.gain.setValueAtTime(0, shimmerAt)
  sg.gain.linearRampToValueAtTime(0.07, shimmerAt + 0.02)
  sg.gain.exponentialRampToValueAtTime(0.001, shimmerAt + 0.18)
  sh.connect(sg)
  sg.connect(audio.destination)
  sh.start(shimmerAt)
  sh.stop(shimmerAt + 0.2)
}

let ctx: AudioCtx | null = null
let lastNavAt = 0

function ensureCtx() {
  if (!ctx) {
    const Ctx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
    if (!Ctx) return null
    ctx = new Ctx()
  }
  return ctx
}

async function unlock() {
  const audio = ensureCtx()
  if (!audio) return null
  if (audio.state === 'suspended') {
    try {
      await audio.resume()
    } catch {
      /* ignore */
    }
  }
  return audio
}

async function play(fn: (audio: AudioCtx) => void) {
  const audio = await unlock()
  if (!audio) return
  fn(audio)
}

export const sfx = {
  unlock,
  mark() {
    return play(markSound)
  },
  unmark() {
    return play(unmarkSound)
  },
  nav() {
    const now = performance.now()
    if (now - lastNavAt < 100) return Promise.resolve()
    lastNavAt = now
    return play(navSound)
  },
  win() {
    return play(winSound)
  },
}
