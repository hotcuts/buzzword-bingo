type CellProps = {
  label: string
  marked: boolean
  index: number
  onMark: (index: number) => void
  onNav: () => void
}

export function Cell({ label, marked, index, onMark, onNav }: CellProps) {
  const isFree = label === 'FREE'
  const className = ['cell', marked || isFree ? 'marked' : '', isFree ? 'free' : '']
    .filter(Boolean)
    .join(' ')

  return (
    <button
      type="button"
      className={className}
      onClick={() => {
        if (!isFree) onMark(index)
      }}
      onMouseEnter={() => {
        if (!isFree) onNav()
      }}
      onFocus={() => {
        if (!isFree) onNav()
      }}
      aria-pressed={marked || isFree}
      aria-label={isFree ? 'Free space' : label}
    >
      <span className="cell-label">{label}</span>
    </button>
  )
}
