import { Tile } from './Tile'
import type { Guess } from '../types/api'
import './GameBoard.css'

interface GameBoardProps {
  guesses: Guess[]
  currentGuess: string
  maxAttempts?: number
  wordLength?: number
}

export function GameBoard({
  guesses,
  currentGuess,
  maxAttempts = 6,
  wordLength = 5,
}: GameBoardProps) {
  const rows: React.ReactNode[] = []

  // Completed guesses
  for (let i = 0; i < guesses.length; i++) {
    const guess = guesses[i]
    rows.push(
      <div key={`guess-${i}`} className="board-row">
        {guess.word.split('').map((letter, j) => (
          <Tile key={j} letter={letter.toUpperCase()} result={guess.results[j]} />
        ))}
      </div>
    )
  }

  // Current guess row
  if (guesses.length < maxAttempts) {
    const currentRow = (
      <div key="current" className="board-row">
        {Array.from({ length: wordLength }).map((_, i) => (
          <Tile
            key={i}
            letter={currentGuess[i] || ''}
            isActive={i === currentGuess.length}
          />
        ))}
      </div>
    )
    rows.push(currentRow)
  }

  // Empty rows
  for (let i = guesses.length + 1; i < maxAttempts; i++) {
    rows.push(
      <div key={`empty-${i}`} className="board-row">
        {Array.from({ length: wordLength }).map((_, j) => (
          <Tile key={j} />
        ))}
      </div>
    )
  }

  return <div className="game-board">{rows}</div>
}
