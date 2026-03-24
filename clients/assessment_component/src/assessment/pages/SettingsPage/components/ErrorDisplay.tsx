import { AlertCircle } from 'lucide-react'

interface ErrorDisplayProps {
  error?: string
}

export const ErrorDisplay = ({ error }: ErrorDisplayProps) => {
  if (!error) return null

  return (
    <div className='flex items-center gap-2 text-red-600 bg-red-50 p-3 rounded-lg'>
      <AlertCircle className='h-4 w-4' />
      <span className='text-sm'>{error}</span>
    </div>
  )
}
