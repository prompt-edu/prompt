import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  ManagementPageHeader,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const ParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const participationsQuery = useGetCoursePhaseParticipants()

  if (!phaseId) return <ErrorPage description='Invalid course phase ID' />

  return (
    <QueryGate
      queries={[participationsQuery]}
      errorFallback={({ refetch }) => (
        <ErrorPage onRetry={refetch} description='Could not fetch participants' />
      )}
    >
      <div id='table-view' className='relative flex flex-col'>
        <ManagementPageHeader>Example Component Participants</ManagementPageHeader>
        <p className='text-sm text-muted-foreground mb-4'>
          This table shows all participants of the Example Component phase.
        </p>
        <div className='w-full'>
          <CoursePhaseParticipationsTable
            phaseId={phaseId}
            participants={participationsQuery.data?.participations ?? []}
          />
        </div>
      </div>
    </QueryGate>
  )
}
