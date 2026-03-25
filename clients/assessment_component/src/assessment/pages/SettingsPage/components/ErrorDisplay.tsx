import { AlertCircle } from 'lucide-react'

interface ErrorDisplayProps {
  error?: string
}

export const ErrorDisplay = ({ error }: ErrorDisplayProps) => {
  if (!error) return null

  return (
    <div className='flex items-center gap-2 rounded-lg bg-destructive/10 p-3 text-destructive'>
      <AlertCircle className='h-4 w-4' />
      <span className='text-sm'>{error}</span>
    </div>
  )
}
