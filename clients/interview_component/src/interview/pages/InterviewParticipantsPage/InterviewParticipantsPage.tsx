import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  ManagementPageHeader,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const InterviewParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const participationsQuery = useGetCoursePhaseParticipants()

  if (!phaseId) return <ErrorPage description='Invalid course phase ID' />

  return (
    <div>
      <ManagementPageHeader>Interview Participants</ManagementPageHeader>
      <QueryGate queries={[participationsQuery]}>
        <CoursePhaseParticipationsTable
          phaseId={phaseId}
          participants={participationsQuery.data?.participations ?? []}
        />
      </QueryGate>
    </div>
  )
}
