import { Loader2 } from 'lucide-react'
import React, { Suspense } from 'react'

const StudentNotesAndHistory = React.lazy(async () => {
  try {
    const module = await import('core/provide')
    return { default: module.StudentNotesAndHistory }
  } catch (err) {
    console.error('[MF] Failed to load student history from core', err)
    return { default: () => null }
  }
})

export const StudentHistoryCard = ({ studentId }: { studentId: string }) => (
  <Suspense
    fallback={
      <div className='flex justify-center p-4'>
        <Loader2 className='h-6 w-6 animate-spin' />
      </div>
    }
  >
    <div className='flex flex-col gap-4'>
      <StudentNotesAndHistory studentId={studentId} />
    </div>
  </Suspense>
)
