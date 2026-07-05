import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  ErrorPage,
  LoadingPage,
} from '@tumaet/prompt-ui-components'
import { TriangleAlert } from 'lucide-react'
import { useParams } from 'react-router-dom'

import { useGetAllTeams } from './hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from './hooks/useGetCoursePhaseConfig'
import { useGetMyEvaluationCompletions } from './hooks/useGetMyEvaluationCompletions'
import { useGetMyEvaluations } from './hooks/useGetMyEvaluations'
import { useGetMyParticipation } from './hooks/useGetMyParticipation'
import { useGetPeerEvaluationCategoriesWithCompetencies } from './hooks/useGetPeerEvaluationCategoriesWithCompetencies'
import { useGetSelfEvaluationCategoriesWithCompetencies } from './hooks/useGetSelfEvaluationCategoriesWithCompetencies'
import { useGetTutorEvaluationCategoriesWithCompetencies } from './hooks/useGetTutorEvaluationCategoriesWithCompetencies'

interface EvaluationDataShellProps {
  children: React.ReactNode
}

export const EvaluationDataShell = ({ children }: EvaluationDataShellProps) => {
  const { isStudentOfCourse } = useCourseStore()
  const { courseId } = useParams<{ courseId: string; phaseId: string }>()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const {
    data: coursePhaseConfig,
    isPending: isCoursePhaseConfigPending,
    isError: isCoursePhaseConfigError,
    refetch: refetchCoursePhaseConfig,
  } = useGetCoursePhaseConfig()

  const {
    isPending: isSelfEvaluationCategoriesPending,
    isError: isSelfEvaluationCategoriesError,
    refetch: refetchSelfEvaluationCategories,
  } = useGetSelfEvaluationCategoriesWithCompetencies(
    coursePhaseConfig?.selfEvaluationEnabled ?? false,
  )

  const {
    isPending: isPeerEvaluationCategoriesPending,
    isError: isPeerEvaluationCategoriesError,
    refetch: refetchPeerEvaluationCategories,
  } = useGetPeerEvaluationCategoriesWithCompetencies(
    coursePhaseConfig?.peerEvaluationEnabled ?? false,
  )

  const {
    isPending: isTutorEvaluationCategoriesPending,
    isError: isTutorEvaluationCategoriesError,
    refetch: refetchTutorEvaluationCategories,
  } = useGetTutorEvaluationCategoriesWithCompetencies(
    coursePhaseConfig?.tutorEvaluationEnabled ?? false,
  )

  const {
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useGetAllTeams()

  const {
    isPending: isParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useGetMyParticipation({ enabled: isStudent })

  const {
    isPending: isCompletionPending,
    isError: isCompletionError,
    refetch: refetchCompletion,
  } = useGetMyEvaluationCompletions({ enabled: isStudent })

  const {
    isPending: isMyEvaluationsPending,
    isError: isMyEvaluationsError,
    refetch: refetchMyEvaluations,
  } = useGetMyEvaluations({ enabled: isStudent })

  const isError =
    (coursePhaseConfig?.selfEvaluationEnabled && isSelfEvaluationCategoriesError) ||
    (coursePhaseConfig?.peerEvaluationEnabled && isPeerEvaluationCategoriesError) ||
    (coursePhaseConfig?.tutorEvaluationEnabled && isTutorEvaluationCategoriesError) ||
    isTeamsError ||
    isCoursePhaseConfigError ||
    (isStudent && isParticipationsError) ||
    (isStudent && isCompletionError) ||
    (isStudent && isMyEvaluationsError)
  const isPending =
    (coursePhaseConfig?.selfEvaluationEnabled && isSelfEvaluationCategoriesPending) ||
    (coursePhaseConfig?.peerEvaluationEnabled && isPeerEvaluationCategoriesPending) ||
    (coursePhaseConfig?.tutorEvaluationEnabled && isTutorEvaluationCategoriesPending) ||
    isTeamsPending ||
    isCoursePhaseConfigPending ||
    (isStudent && isParticipationsPending) ||
    (isStudent && isCompletionPending) ||
    (isStudent && isMyEvaluationsPending)
  const refetch = () => {
    if (coursePhaseConfig?.selfEvaluationEnabled) {
      refetchSelfEvaluationCategories()
    }
    if (coursePhaseConfig?.peerEvaluationEnabled) {
      refetchPeerEvaluationCategories()
    }
    if (coursePhaseConfig?.tutorEvaluationEnabled) {
      refetchTutorEvaluationCategories()
    }
    refetchTeams()
    refetchCoursePhaseConfig()
    if (isStudent) {
      refetchCoursePhaseParticipations()
      refetchCompletion()
      refetchMyEvaluations()
    }
  }

  if (isError)
    return (
      <ErrorPage
        onRetry={refetch}
        description='Could not fetch self, peer, or tutor evaluation categories'
      />
    )
  if (isPending) return <LoadingPage />

  return (
    <>
      {!isStudent && (
        <Alert>
          <TriangleAlert className='h-4 w-4' />
          <AlertTitle>Your are not a student of this course.</AlertTitle>
          <AlertDescription>
            The following components are disabled because you are not a student of this course.
            Evaluations for self and peer assessments are currently only available for students. The
            platform will show a random team regardless, to demonstrate the functionality.
          </AlertDescription>
        </Alert>
      )}
      {children}
    </>
  )
}
