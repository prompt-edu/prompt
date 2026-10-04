import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  LoadingPage,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const ParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: coursePhaseParticipations,
    isPending,
    isError,
    refetch,
  } = useGetCoursePhaseParticipants()

  if (isError) return <ErrorPage onRetry={refetch} description='Could not fetch participants' />
  if (isPending) return <LoadingPage />

  return (
    <div id='table-view' className='relative flex flex-col'>
      <ManagementPageHeader>Example Component Participants</ManagementPageHeader>
      <p className='text-sm text-muted-foreground mb-4'>
        This table shows all participants of the Example Component phase.
      </p>
      <div className='w-full'>
        <CoursePhaseParticipationsTable
          phaseId={phaseId!}
          participants={coursePhaseParticipations.participations ?? []}
        />
      </div>
    </div>
  )
}
