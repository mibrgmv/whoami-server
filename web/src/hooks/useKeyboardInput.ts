import { useEffect, useCallback } from 'react'

interface UseKeyboardInputProps {
  onLetter: (letter: string) => void
  onEnter: () => void
  onBackspace: () => void
  disabled: boolean
}

export function useKeyboardInput({ onLetter, onEnter, onBackspace, disabled }: UseKeyboardInputProps) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (disabled) return

      if (e.key === 'Enter') {
        e.preventDefault()
        onEnter()
      } else if (e.key === 'Backspace') {
        e.preventDefault()
        onBackspace()
      } else if (/^[a-zA-Z]$/.test(e.key)) {
        onLetter(e.key)
      }
    },
    [disabled, onEnter, onBackspace, onLetter]
  )

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])
}
