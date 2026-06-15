import {
  DeletionSubrequestStatus,
  type PrivacyDeletionSubrequest,
} from '@core/network/queries/privacyStudentDataDeletion'
import { CircleCheck, CircleDashed, CircleX, Loader2 } from 'lucide-react'

interface PrivacyDeletionSubrequestListProps {
  subrequests: PrivacyDeletionSubrequest[]
}

const subrequestStatusLabel: Record<DeletionSubrequestStatus, string> = {
  [DeletionSubrequestStatus.pending]: 'Pending',
  [DeletionSubrequestStatus.in_progress]: 'In progress',
  [DeletionSubrequestStatus.succeeded]: 'Succeeded',
  [DeletionSubrequestStatus.failed]: 'Failed',
}

function SubrequestStatusIcon({ status }: { status: DeletionSubrequestStatus }) {
  const className = 'h-4 w-4'
  switch (status) {
    case DeletionSubrequestStatus.pending:
      return <CircleDashed className={`${className} text-muted-foreground`} />
    case DeletionSubrequestStatus.in_progress:
      return <Loader2 className={`${className} animate-spin text-muted-foreground`} />
    case DeletionSubrequestStatus.succeeded:
      return <CircleCheck className={`${className} text-green-600 dark:text-green-400`} />
    case DeletionSubrequestStatus.failed:
      return <CircleX className={`${className} text-red-600 dark:text-red-400`} />
  }
}

export function PrivacyDeletionSubrequestList({ subrequests }: PrivacyDeletionSubrequestListProps) {
  if (subrequests.length === 0) return null

  return (
    <div className='rounded-lg border border-border'>
      <div className='px-4 py-2 border-b border-border'>
        <p className='text-sm font-medium'>Results</p>
      </div>
      <ul className='divide-y divide-border'>
        {subrequests.map((sub) => (
          <li key={sub.id} className='px-4 py-3 flex items-center gap-3'>
            <SubrequestStatusIcon status={sub.status} />
            <div className='flex-1 min-w-0'>
              <p className='text-sm font-medium'>{sub.source_name}</p>
              <p className='text-xs text-muted-foreground'>{subrequestStatusLabel[sub.status]}</p>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
