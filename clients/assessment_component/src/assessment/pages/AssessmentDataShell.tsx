import { ErrorPage, LoadingPage } from '@tumaet/prompt-ui-components'
import { AssessmentType } from '../interfaces/assessmentType'
import { useGetAllCategoriesWithCompetencies } from './hooks/useGetAllCategoriesWithCompetencies'
import { useGetAllScoreLevels } from './hooks/useGetAllScoreLevels'
import { useGetAllTeams } from './hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from './hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseParticipations } from './hooks/useGetCoursePhaseParticipations'
import { useGetEvaluationCategoriesWithCompetencies } from './hooks/useGetEvaluationCategoriesWithCompetencies'

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
  } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.SELF,
    coursePhaseConfig?.selfEvaluationEnabled ?? false,
  )

  const {
    isPending: isPeerEvaluationCategoriesPending,
    isError: isPeerEvaluationCategoriesError,
    refetch: refetchPeerEvaluationCategories,
  } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.PEER,
    coursePhaseConfig?.peerEvaluationEnabled ?? false,
  )

  const {
    isPending: isTutorEvaluationCategoriesPending,
    isError: isTutorEvaluationCategoriesError,
    refetch: refetchTutorEvaluationCategories,
  } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.TUTOR,
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
