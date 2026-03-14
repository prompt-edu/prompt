import { ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { CoursePhaseParticipationsTable } from '@/components/pages/CoursePhaseParticipationsTable/CoursePhaseParticipationsTable'
import { Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { useGetCoursePhaseParticipants } from '@/hooks/useGetCoursePhaseParticipants'

export const IntroCourseParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: coursePhaseParticipations,
    isPending,
    isError,
    refetch,
  } = useGetCoursePhaseParticipants()

  return (
    <div>
      {isError ? (
        <ErrorPage onRetry={refetch} />
      ) : isPending ? (
        <div className='flex justify-center items-center h-64'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
        </div>
      ) : (
        <>
          <ManagementPageHeader>Intro Course Participants</ManagementPageHeader>
          <CoursePhaseParticipationsTable
            phaseId={phaseId!}
            participants={coursePhaseParticipations?.participations ?? []}
          />
        </>
      )}
    </div>
  )
}
