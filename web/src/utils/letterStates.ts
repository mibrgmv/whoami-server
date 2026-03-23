import type { Guess, LetterResult } from '../types/api'
import { LetterResultValues } from '../types/api'

export function updateLetterStates(
  current: Record<string, LetterResult>,
  guess: Guess
): Record<string, LetterResult> {
  const updated = { ...current }
  const word = guess.word.toUpperCase()

  for (let i = 0; i < word.length; i++) {
    const letter = word[i]
    const result = guess.results[i]

    if (result === LetterResultValues.CORRECT) {
      updated[letter] = LetterResultValues.CORRECT
    } else if (result === LetterResultValues.PRESENT && updated[letter] !== LetterResultValues.CORRECT) {
      updated[letter] = LetterResultValues.PRESENT
    } else if (result === LetterResultValues.ABSENT && !updated[letter]) {
      updated[letter] = LetterResultValues.ABSENT
    }
  }

  return updated
}
