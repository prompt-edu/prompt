import { useParams } from 'react-router-dom'
import { Loader2 } from 'lucide-react'
import { ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { CoursePhaseParticipationsTable } from '@tumaet/prompt-ui-components'
import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'

export const MatchingParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: coursePhaseParticipations,
    isPending,
    isError,
    refetch,
  } = useGetCoursePhaseParticipants()

  return (
    <div>
      <ManagementPageHeader>Matching Participants</ManagementPageHeader>
      {isError ? (
        <ErrorPage onRetry={refetch} />
      ) : isPending ? (
        <div className='flex justify-center items-center h-64'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
        </div>
      ) : (
        <CoursePhaseParticipationsTable
          phaseId={phaseId!}
          participants={coursePhaseParticipations.participations ?? []}
          exportDeps={{ prevDataKeys: ['score'] }}
        />
      )}
    </div>
  )
}
