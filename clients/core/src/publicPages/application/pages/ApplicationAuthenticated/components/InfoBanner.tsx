import { AlertCircle, Info } from 'lucide-react'
import type React from 'react'

interface InfoBannerProps {
  status: 'not_applied' | 'new_user' | 'applied' | string
}

export const InfoBanner: React.FC<InfoBannerProps> = ({ status }) => {
  const getBannerMessage = () => {
    if (status === 'not_applied' || status === 'new_user') {
      return 'Your changes are not saved until you submit. You can resubmit as often as you want until the deadline.'
    } else if (status === 'applied') {
      return 'You have already successfully submitted. You can resubmit as often as you want until the deadline, but your old answers will be overwritten.'
    }
    return ''
  }

  const message = getBannerMessage()

  if (!message) return null

  const isApplied = status === 'applied'

  return (
    <div
      className={`border-l-4 p-4 mb-4 ${
        isApplied
          ? 'bg-yellow-100 dark:bg-yellow-900/30 border-yellow-500 text-yellow-700 dark:text-yellow-300'
          : 'bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300'
      }`}
      role='alert'
    >
      <div className='flex items-center'>
        {isApplied ? (
          <AlertCircle className='h-5 w-5 mr-2 text-yellow-600 dark:text-yellow-400' />
        ) : (
          <Info className='h-5 w-5 mr-2 text-gray-500 dark:text-gray-400' />
        )}
        <p className={`${isApplied ? 'font-medium' : 'font-normal'}`}>{message}</p>
      </div>
    </div>
  )
}
