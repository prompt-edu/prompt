import { ErrorPage, LoadingPage } from '@tumaet/prompt-ui-components'
import { useGetAllCategoriesWithCompetencies } from './hooks/useGetAllCategoriesWithCompetencies'
import { useGetAllScoreLevels } from './hooks/useGetAllScoreLevels'
import { useGetAllTeams } from './hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from './hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseParticipations } from './hooks/useGetCoursePhaseParticipations'
import { useGetPeerEvaluationCategoriesWithCompetencies } from './hooks/useGetPeerEvaluationCategoriesWithCompetencies'
import { useGetSelfEvaluationCategoriesWithCompetencies } from './hooks/useGetSelfEvaluationCategoriesWithCompetencies'
import { useGetTutorEvaluationCategoriesWithCompetencies } from './hooks/useGetTutorEvaluationCategoriesWithCompetencies'

interface AssessmentDataShellProps {
  children: React.ReactNode
}

export const AssessmentDataShell = ({ children }: AssessmentDataShellProps) => {
  const {
    isPending: isCoursePhaseParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useGetCoursePhaseParticipations()

  const {
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useGetAllTeams()

  const {
    isPending: isCategoriesPending,
    isError: isCategoriesError,
    refetch: refetchCategories,
  } = useGetAllCategoriesWithCompetencies()

  const {
    isPending: isScoreLevelsPending,
    isError: isScoreLevelsError,
    refetch: refetchScoreLevels,
  } = useGetAllScoreLevels()

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

  const isError =
    isParticipationsError ||
    isTeamsError ||
    isCategoriesError ||
    isScoreLevelsError ||
    isCoursePhaseConfigError ||
    (coursePhaseConfig?.selfEvaluationEnabled && isSelfEvaluationCategoriesError) ||
    (coursePhaseConfig?.peerEvaluationEnabled && isPeerEvaluationCategoriesError) ||
    (coursePhaseConfig?.tutorEvaluationEnabled && isTutorEvaluationCategoriesError)
  const isPending =
    isCoursePhaseParticipationsPending ||
    isTeamsPending ||
    isCategoriesPending ||
    isScoreLevelsPending ||
    isCoursePhaseConfigPending ||
    (coursePhaseConfig?.selfEvaluationEnabled && isSelfEvaluationCategoriesPending) ||
    (coursePhaseConfig?.peerEvaluationEnabled && isPeerEvaluationCategoriesPending) ||
    (coursePhaseConfig?.tutorEvaluationEnabled && isTutorEvaluationCategoriesPending)

  const refetch = () => {
    refetchTeams()
    refetchCoursePhaseParticipations()
    refetchCategories()
    refetchScoreLevels()
    refetchCoursePhaseConfig()
    if (coursePhaseConfig?.selfEvaluationEnabled) {
      refetchSelfEvaluationCategories()
    }
    if (coursePhaseConfig?.peerEvaluationEnabled) {
      refetchPeerEvaluationCategories()
    }
    if (coursePhaseConfig?.tutorEvaluationEnabled) {
      refetchTutorEvaluationCategories()
    }
  }

  return <>{isError ? <ErrorPage onRetry={refetch} /> : isPending ? <LoadingPage /> : children}</>
}
