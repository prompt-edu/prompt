import { ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { CoursePhaseParticipationsTable } from '@tumaet/prompt-ui-components'
import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'

export const ParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: coursePhaseParticipations,
    isPending,
    isError,
    refetch,
  } = useGetCoursePhaseParticipants()

  if (isError) return <ErrorPage onRetry={refetch} description='Could not fetch participants' />
  if (isPending)
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )

  return (
    <div id='table-view' className='relative flex flex-col'>
      <ManagementPageHeader>Template Component Participants</ManagementPageHeader>
      <p className='text-sm text-muted-foreground mb-4'>
        This table shows all participants of the Template Component phase.
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
