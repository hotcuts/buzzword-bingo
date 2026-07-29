import { Cell } from './Cell'

const COLUMNS = ['B', 'I', 'N', 'G', 'O'] as const

type BingoBoardProps = {
  cells: string[]
  marked: boolean[]
  won: boolean
  onMark: (index: number) => void
  onNav: () => void
  onPlayAgain: () => void
}

export function BingoBoard({ cells, marked, won, onMark, onNav, onPlayAgain }: BingoBoardProps) {
  return (
    <div className="board-wrap">
      <div className="columns" aria-hidden="true">
        {COLUMNS.map((letter) => (
          <span key={letter}>{letter}</span>
        ))}
      </div>
      <div className="board" aria-label="Bingo board">
        {cells.map((label, index) => (
          <Cell
            key={`${index}-${label}`}
            label={label}
            marked={marked[index]}
            index={index}
            onMark={onMark}
            onNav={onNav}
          />
        ))}
      </div>

      {won && (
        <div className="win-overlay">
          <div className="win-card">
            <p className="win-title">Bingo!</p>
            <button type="button" className="play-again-btn" onClick={onPlayAgain}>
              Play again
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
