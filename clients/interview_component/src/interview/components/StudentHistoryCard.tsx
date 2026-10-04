/// <reference path="../../declaration.d.ts" />
import { Loader2 } from 'lucide-react'
import React, { Suspense } from 'react'
import { ErrorBoundary } from './ErrorBoundary'

const StudentNotesAndHistory = React.lazy(() =>
  import('core/provide').then((module) => ({ default: module.StudentNotesAndHistory })),
)

export const StudentHistoryCard = ({ studentId }: { studentId: string }) => (
  <ErrorBoundary fallback={null}>
    <div
      data-student-history
      className='flex flex-col gap-4 lg:col-span-2 lg:flex-row lg:gap-8 lg:*:min-w-0 lg:*:flex-1 xl:col-span-1 xl:flex-col xl:gap-4 xl:*:flex-none'
    >
      <Suspense
        fallback={
          <div className='flex justify-center p-4'>
            <Loader2 className='h-6 w-6 animate-spin' />
          </div>
        }
      >
        <StudentNotesAndHistory studentId={studentId} />
      </Suspense>
    </div>
  </ErrorBoundary>
)
