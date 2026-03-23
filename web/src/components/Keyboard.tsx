import type { LetterResult } from '../types/api'
import { LetterResultValues } from '../types/api'
import './Keyboard.css'

interface KeyboardProps {
  onKey: (key: string) => void
  onEnter: () => void
  onBackspace: () => void
  letterStates: Record<string, LetterResult>
  disabled?: boolean
}

const ROWS = [
  ['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
  ['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'],
  ['ENTER', 'Z', 'X', 'C', 'V', 'B', 'N', 'M', 'BACKSPACE'],
]

function getStateClass(state: LetterResult | undefined): string {
  if (!state) return ''
  switch (state) {
    case LetterResultValues.CORRECT:
      return 'key--correct'
    case LetterResultValues.PRESENT:
      return 'key--present'
    case LetterResultValues.ABSENT:
      return 'key--absent'
    default:
      return ''
  }
}

export function Keyboard({
  onKey,
  onEnter,
  onBackspace,
  letterStates,
  disabled,
}: KeyboardProps) {
  const handleClick = (key: string) => {
    if (disabled) return

    if (key === 'ENTER') {
      onEnter()
    } else if (key === 'BACKSPACE') {
      onBackspace()
    } else {
      onKey(key)
    }
  }

  return (
    <div className="keyboard">
      {ROWS.map((row, i) => (
        <div key={i} className="keyboard-row">
          {row.map((key) => {
            const state = letterStates[key]
            const isWide = key === 'ENTER' || key === 'BACKSPACE'

            return (
              <button
                key={key}
                className={`key ${getStateClass(state)} ${isWide ? 'key--wide' : ''}`}
                onClick={() => handleClick(key)}
                disabled={disabled}
              >
                {key === 'BACKSPACE' ? '⌫' : key}
              </button>
            )
          })}
        </div>
      ))}
    </div>
  )
}
