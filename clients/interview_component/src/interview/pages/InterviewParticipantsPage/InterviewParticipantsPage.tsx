import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  LoadingPage,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const InterviewParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: coursePhaseParticipations,
    isPending,
    isError,
    refetch,
  } = useGetCoursePhaseParticipants()

  return (
    <div>
      <ManagementPageHeader>Interview Participants</ManagementPageHeader>
      {isError ? (
        <ErrorPage onRetry={refetch} />
      ) : isPending ? (
        <LoadingPage />
      ) : (
        <CoursePhaseParticipationsTable
          phaseId={phaseId!}
          participants={coursePhaseParticipations.participations ?? []}
        />
      )}
    </div>
  )
}
