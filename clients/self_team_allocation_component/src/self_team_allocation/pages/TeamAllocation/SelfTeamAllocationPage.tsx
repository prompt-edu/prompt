import { useQuery } from '@tanstack/react-query'
import {
  type CoursePhaseParticipationWithStudent,
  getOwnCoursePhaseParticipation,
  type Team,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  ErrorPage,
  UnauthorizedPage,
} from '@tumaet/prompt-ui-components'
import { Loader2, TriangleAlert } from 'lucide-react'
import { useParams } from 'react-router-dom'
import type { Timeframe } from '../../interfaces/timeframe'
import { getAllTeams } from '../../network/queries/getAllTeams'
import { getTimeframe } from '../../network/queries/getSurveyTimeframe'
import { TeamSelection } from './components/TeamSelection'

export const SelfTeamAllocationPage = () => {
  const { isStudentOfCourse } = useCourseStore()
  const { courseId = '', phaseId = '' } = useParams<{ courseId: string; phaseId: string }>()
  const isStudent = isStudentOfCourse(courseId)

  const {
    data: participation,
    isPending: isParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
    error: participationError,
  } = useQuery<CoursePhaseParticipationWithStudent>({
    queryKey: ['course_phase_participation', phaseId],
    queryFn: () => getOwnCoursePhaseParticipation(phaseId),
    enabled: isStudent,
  })

  const {
    data: teams,
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useQuery<Team[]>({
    queryKey: ['self_team_allocations', phaseId],
    queryFn: () => getAllTeams(phaseId),
  })

  const {
    data: timeframe,
    isPending: isTimeframePending,
    isError: isTimeframeError,
    refetch: refetchTimeframe,
  } = useQuery<Timeframe>({
    queryKey: ['timeframe', phaseId],
    queryFn: () => getTimeframe(phaseId),
  })

  const isError = isParticipationsError || isTeamsError || isTimeframeError
  const isPending = isParticipationsPending || isTeamsPending || isTimeframePending
  const refetch = () => {
    refetchCoursePhaseParticipations()
    refetchTeams()
    refetchTimeframe()
  }

  if (isTeamsPending || (isStudent && isPending)) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  if (isStudent && isError) {
    if (participationError && participationError.message.includes('404')) {
      return <UnauthorizedPage backUrl={`/management/course/${courseId}`} />
    }
    return <ErrorPage onRetry={refetch} />
  }

  const cpId = participation?.courseParticipationID

  return (
    <>
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
          courseParticipationID={cpId}
          refetchTeams={refetchTeams}
          timeframe={timeframe}
        />
      )}
    </>
  )
}
