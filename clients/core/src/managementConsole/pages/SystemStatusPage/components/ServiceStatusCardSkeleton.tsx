import { Skeleton } from '@tumaet/prompt-ui-components'

export function ServiceStatusCardSkeleton() {
  return (
    <div className='flex flex-col gap-2'>
      <Skeleton className='h-4 w-1/2' />
      <Skeleton className='h-4 w-1/3' />
      <Skeleton className='h-px w-full' />
      <Skeleton className='h-4 w-full' />
      <Skeleton className='h-4 w-3/4' />
      <Skeleton className='h-4 w-5/6' />
      <Skeleton className='h-4 w-2/3' />
    </div>
  )
}
