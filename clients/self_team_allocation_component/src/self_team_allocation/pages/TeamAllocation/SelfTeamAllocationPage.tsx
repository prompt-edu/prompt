import { useQuery } from '@tanstack/react-query'
import {
  type CoursePhaseParticipationWithStudent,
  EDITOR_ROLES,
  getOwnCoursePhaseParticipation,
  getPermissionString,
  type Team,
  useAuthStore,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  ErrorPage,
  QueryGate,
  UnauthorizedPage,
} from '@tumaet/prompt-ui-components'
import { TriangleAlert } from 'lucide-react'
import { useParams } from 'react-router-dom'
import type { Timeframe } from '../../interfaces/timeframe'
import { getAllTeams } from '../../network/queries/getAllTeams'
import { getTimeframe } from '../../network/queries/getSurveyTimeframe'
import { TeamSelection } from './components/TeamSelection'

export const SelfTeamAllocationPage = () => {
  const { courses, isStudentOfCourse } = useCourseStore()
  const { permissions } = useAuthStore()
  const { courseId = '', phaseId = '' } = useParams<{ courseId: string; phaseId: string }>()
  const course = courses.find((c) => c.id === courseId)
  const isManager = EDITOR_ROLES.some((role) =>
    permissions.includes(getPermissionString(role, course?.name, course?.semesterTag)),
  )
  const isStudent = isStudentOfCourse(courseId) && !isManager

  const participationQuery = useQuery<CoursePhaseParticipationWithStudent>({
    queryKey: ['course_phase_participation', phaseId],
    queryFn: () => getOwnCoursePhaseParticipation(phaseId),
    enabled: isStudent,
  })

  const teamsQuery = useQuery<Team[]>({
    queryKey: ['self_team_allocations', phaseId],
    queryFn: () => getAllTeams(phaseId),
  })

  const timeframeQuery = useQuery<Timeframe>({
    queryKey: ['timeframe', phaseId],
    queryFn: () => getTimeframe(phaseId),
  })

  const { data: participation, error: participationError } = participationQuery
  const teams = teamsQuery.data
  const timeframe = timeframeQuery.data

  return (
    <QueryGate
      queries={
        isStudent ? [participationQuery, teamsQuery, timeframeQuery] : [teamsQuery, timeframeQuery]
      }
      errorFallback={({ refetch }) =>
        participationError?.message.includes('404') ? (
          <UnauthorizedPage backUrl={`/management/course/${courseId}`} />
        ) : (
          <ErrorPage onRetry={refetch} />
        )
      }
    >
      {!isStudent && (
        <Alert>
          <TriangleAlert className='h-4 w-4' />
          <AlertTitle>You are not a student of this course.</AlertTitle>
          <AlertDescription>
            The team-allocation UI is disabled because you’re not enrolled.
          </AlertDescription>
        </Alert>
      )}

      {teams && timeframe && (
        <TeamSelection
          teams={teams}
          courseParticipationID={participation?.courseParticipationID}
          refetchTeams={teamsQuery.refetch}
          timeframe={timeframe}
        />
      )}
    </QueryGate>
  )
}
