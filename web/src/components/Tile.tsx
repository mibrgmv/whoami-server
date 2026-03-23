import type { LetterResult } from '../types/api'
import { LetterResultValues } from '../types/api'
import './Tile.css'

interface TileProps {
  letter?: string
  result?: LetterResult
  isActive?: boolean
}

function getResultClass(result: LetterResult | undefined): string {
  if (!result) return ''
  switch (result) {
    case LetterResultValues.CORRECT:
      return 'tile--correct'
    case LetterResultValues.PRESENT:
      return 'tile--present'
    case LetterResultValues.ABSENT:
      return 'tile--absent'
    default:
      return ''
  }
}

export function Tile({ letter, result, isActive }: TileProps) {
  const className = [
    'tile',
    getResultClass(result),
    isActive && 'tile--active',
    letter && !result && 'tile--filled',
  ]
    .filter(Boolean)
    .join(' ')

  return <div className={className}>{letter}</div>
}
