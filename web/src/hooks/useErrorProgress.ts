import { useEffect, useState } from 'react'

export function useErrorProgress(error: string | null, duration: number, onClear: () => void) {
  const [progress, setProgress] = useState(100)

  useEffect(() => {
    if (error) {
      setProgress(100)

      const intervalTime = 50
      const decrement = (intervalTime / duration) * 100

      const progressInterval = setInterval(() => {
        setProgress((prev) => {
          const next = prev - decrement
          if (next <= 0) {
            clearInterval(progressInterval)
            return 0
          }
          return next
        })
      }, intervalTime)

      const timeout = setTimeout(() => {
        onClear()
      }, duration)

      return () => {
        clearTimeout(timeout)
        clearInterval(progressInterval)
      }
    }
  }, [error, duration, onClear])

  return progress
}
