import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ManagementPageHeader,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const InterviewParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const participationsQuery = useGetCoursePhaseParticipants()

  return (
    <div>
      <ManagementPageHeader>Interview Participants</ManagementPageHeader>
      <QueryGate queries={[participationsQuery]}>
        <CoursePhaseParticipationsTable
          phaseId={phaseId!}
          participants={participationsQuery.data?.participations ?? []}
        />
      </QueryGate>
    </div>
  )
}
