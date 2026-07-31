import type { ResetPeriod } from './api'

export function formatDate(iso: string): string {
  const d = new Date(`${iso}T12:00:00`)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

/** Monday of the given ISO week (year + week number). */
export function mondayOfISOWeek(year: number, week: number): Date {
  const jan4 = new Date(year, 0, 4, 12, 0, 0)
  const day = jan4.getDay() || 7 // Mon=1 … Sun=7
  const mondayWeek1 = new Date(jan4)
  mondayWeek1.setDate(jan4.getDate() - day + 1)
  const monday = new Date(mondayWeek1)
  monday.setDate(mondayWeek1.getDate() + (week - 1) * 7)
  return monday
}

export function formatPeriodLabel(periodKey: string, period: ResetPeriod = 'daily'): string {
  const weekMatch = periodKey.match(/^(\d{4})-W(\d{2})$/)
  if (period === 'weekly' || weekMatch) {
    if (!weekMatch) return periodKey
    const year = Number(weekMatch[1])
    const week = Number(weekMatch[2])
    const monday = mondayOfISOWeek(year, week)
    if (Number.isNaN(monday.getTime())) return periodKey
    const label = monday.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
    return `Week of ${label}`
  }
  return formatDate(periodKey)
}

export function hasSelections(cells: string[], marked: boolean[]): boolean {
  return marked.some((m, i) => m && cells[i] !== 'FREE')
}

export function parseTermsText(raw: string): string[] {
  return raw
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => line !== '' && !line.startsWith('#'))
}
