import type { LetterResult } from '../types/api'
import './Tile.css'

interface TileProps {
  letter?: string
  result?: LetterResult
  isActive?: boolean
}

export function Tile({ letter, result, isActive }: TileProps) {
  const className = [
    'tile',
    result && `tile--${result}`,
    isActive && 'tile--active',
    letter && !result && 'tile--filled',
  ]
    .filter(Boolean)
    .join(' ')

  return <div className={className}>{letter}</div>
}
