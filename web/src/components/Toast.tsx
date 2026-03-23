import { useEffect, useState } from 'react'
import { useToastStore, Toast } from '../stores/toastStore'
import './Toast.css'

function ToastItem({ toast, onRemove }: { toast: Toast; onRemove: () => void }) {
  const [progress, setProgress] = useState(100)

  useEffect(() => {
    const intervalTime = 50
    const decrement = (intervalTime / toast.duration) * 100

    const interval = setInterval(() => {
      setProgress((prev) => {
        const next = prev - decrement
        return next <= 0 ? 0 : next
      })
    }, intervalTime)

    return () => clearInterval(interval)
  }, [toast.duration])

  return (
    <div className={`toast toast-${toast.type}`} onClick={onRemove}>
      <span className="toast-message">{toast.message}</span>
      <div className="toast-progress" style={{ width: `${progress}%` }} />
    </div>
  )
}

export function ToastContainer() {
  const { toasts, removeToast } = useToastStore()

  if (toasts.length === 0) return null

  return (
    <div className="toast-container">
      {toasts.map((toast) => (
        <ToastItem
          key={toast.id}
          toast={toast}
          onRemove={() => removeToast(toast.id)}
        />
      ))}
    </div>
  )
}
