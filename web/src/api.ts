export type ResetPeriod = 'daily' | 'weekly'

export type GameState = {
  name: string
  date: string
  cells: string[]
  marked: boolean[]
  won: boolean
  winCount: number
  period: ResetPeriod
}

export type TermsInfo = {
  terms: string[]
  custom: boolean
  count: number
}

async function readError(res: Response): Promise<string> {
  const text = (await res.text()).trim()
  return text || res.statusText || 'request failed'
}

async function json<T>(res: Response): Promise<T> {
  if (!res.ok) throw new Error(await readError(res))
  return res.json() as Promise<T>
}

export async function fetchState(): Promise<GameState> {
  return json(await fetch('/api/state'))
}

export async function markCell(index: number): Promise<GameState> {
  return json(
    await fetch('/api/mark', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ index }),
    }),
  )
}

export async function resetBoard(): Promise<GameState> {
  return json(await fetch('/api/reset', { method: 'POST' }))
}

export async function resetAll(): Promise<GameState> {
  return json(await fetch('/api/reset-all', { method: 'POST' }))
}

export async function fetchTerms(): Promise<TermsInfo> {
  return json(await fetch('/api/terms'))
}

export async function addTerms(terms: string[]): Promise<TermsInfo> {
  return json(
    await fetch('/api/terms/add', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ terms }),
    }),
  )
}

export async function removeTerms(terms: string[]): Promise<TermsInfo> {
  return json(
    await fetch('/api/terms/remove', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ terms }),
    }),
  )
}

export async function replaceTerms(terms: string[]): Promise<TermsInfo> {
  return json(
    await fetch('/api/terms', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ terms }),
    }),
  )
}

export async function resetTerms(): Promise<TermsInfo> {
  return json(await fetch('/api/terms/reset', { method: 'POST' }))
}

export async function setName(name: string): Promise<GameState> {
  return json(
    await fetch('/api/name', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    }),
  )
}

export async function setPeriod(period: ResetPeriod): Promise<GameState> {
  return json(
    await fetch('/api/period', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ period }),
    }),
  )
}
