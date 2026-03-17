import { Check, ChevronsRight, Clock, X } from 'lucide-react'
import { PassStatus } from '@tumaet/prompt-shared-state'
import { ReactElement } from 'react'

function RoundLayout({
  children,
  className,
}: {
  children: ReactElement
  className: string
}): ReactElement {
  return (
    <div
      className={
        'w-7 h-7 overflow-hidden rounded-full flex items-center justify-center ' + className
      }
    >
      {children}
    </div>
  )
}

export function ProgressIndicator({
  passStatus,
}: {
  passStatus: PassStatus | 'CURRENT'
}): ReactElement {
  return (
    <>
      {passStatus == PassStatus.PASSED && (
        <RoundLayout className='bg-green-100 dark:bg-green-900'>
          <Check className='w-4 h-4 text-green-800 dark:text-green-300' />
        </RoundLayout>
      )}
      {passStatus == PassStatus.FAILED && (
        <RoundLayout className='bg-red-100 dark:bg-red-900'>
          <X className='w-4 h-4 text-red-800 dark:text-red-300' />
        </RoundLayout>
      )}
      {passStatus == PassStatus.NOT_ASSESSED && (
        <RoundLayout className='bg-gray-100 dark:bg-gray-900'>
          <Clock className='w-4 h-4 text-gray-800 dark:text-gray-300' />
        </RoundLayout>
      )}
      {passStatus == 'CURRENT' && (
        <RoundLayout className='bg-blue-100 dark:bg-blue-900'>
          {/* <img src='/prompt_logo.svg' alt='Prompt logo' className='size-6' /> */}
          <ChevronsRight className='w-5 h-5 text-blue-800 dark:text-blue-300' />
        </RoundLayout>
      )}
    </>
  )
}
