import { ErrorPage, LoadingPage } from '@tumaet/prompt-ui-components'
import { useEffect } from 'react'
import { useCategoryStore } from '../zustand/useCategoryStore'
import { useCoursePhaseConfigStore } from '../zustand/useCoursePhaseConfigStore'
import { useParticipationStore } from '../zustand/useParticipationStore'
import { usePeerEvaluationCategoryStore } from '../zustand/usePeerEvaluationCategoryStore'
import { useScoreLevelStore } from '../zustand/useScoreLevelStore'
import { useSelfEvaluationCategoryStore } from '../zustand/useSelfEvaluationCategoryStore'
import { useTeamStore } from '../zustand/useTeamStore'
import { useTutorEvaluationCategoryStore } from '../zustand/useTutorEvaluationCategoryStore'
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
  const { setParticipations } = useParticipationStore()
  const { setTeams } = useTeamStore()
  const { setCategories } = useCategoryStore()
  const { setScoreLevels } = useScoreLevelStore()
  const { setCoursePhaseConfig } = useCoursePhaseConfigStore()
  const { setSelfEvaluationCategories } = useSelfEvaluationCategoryStore()
  const { setPeerEvaluationCategories } = usePeerEvaluationCategoryStore()
  const { setTutorEvaluationCategories } = useTutorEvaluationCategoryStore()

  const {
    data: coursePhaseParticipations,
    isPending: isCoursePhaseParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useGetCoursePhaseParticipations()

  const {
    data: teams,
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useGetAllTeams()

  const {
    data: categories,
    isPending: isCategoriesPending,
    isError: isCategoriesError,
    refetch: refetchCategories,
  } = useGetAllCategoriesWithCompetencies()

  const {
    data: scoreLevels,
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
    data: selfEvaluationCategories,
    isPending: isSelfEvaluationCategoriesPending,
    isError: isSelfEvaluationCategoriesError,
    refetch: refetchSelfEvaluationCategories,
  } = useGetSelfEvaluationCategoriesWithCompetencies(
    coursePhaseConfig?.selfEvaluationEnabled ?? false,
  )

  const {
    data: peerEvaluationCategories,
    isPending: isPeerEvaluationCategoriesPending,
    isError: isPeerEvaluationCategoriesError,
    refetch: refetchPeerEvaluationCategories,
  } = useGetPeerEvaluationCategoriesWithCompetencies(
    coursePhaseConfig?.peerEvaluationEnabled ?? false,
  )

  const {
    data: tutorEvaluationCategories,
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

  useEffect(() => {
    if (coursePhaseParticipations) {
      setParticipations(coursePhaseParticipations)
    }
  }, [coursePhaseParticipations, setParticipations])

  useEffect(() => {
    if (teams) {
      setTeams(teams)
    }
  }, [teams, setTeams])

  useEffect(() => {
    if (categories) {
      setCategories(categories)
    }
  }, [categories, setCategories])

  useEffect(() => {
    if (scoreLevels) {
      setScoreLevels(scoreLevels)
    }
  }, [scoreLevels, setScoreLevels])

  useEffect(() => {
    if (coursePhaseConfig) {
      setCoursePhaseConfig(coursePhaseConfig)
    }
  }, [coursePhaseConfig, setCoursePhaseConfig])

  useEffect(() => {
    if (coursePhaseConfig?.selfEvaluationEnabled && selfEvaluationCategories) {
      setSelfEvaluationCategories(selfEvaluationCategories)
    }
  }, [
    coursePhaseConfig?.selfEvaluationEnabled,
    selfEvaluationCategories,
    setSelfEvaluationCategories,
  ])

  useEffect(() => {
    if (coursePhaseConfig?.peerEvaluationEnabled && peerEvaluationCategories) {
      setPeerEvaluationCategories(peerEvaluationCategories)
    }
  }, [
    coursePhaseConfig?.peerEvaluationEnabled,
    peerEvaluationCategories,
    setPeerEvaluationCategories,
  ])

  useEffect(() => {
    if (coursePhaseConfig?.tutorEvaluationEnabled && tutorEvaluationCategories) {
      setTutorEvaluationCategories(tutorEvaluationCategories)
    }
  }, [
    coursePhaseConfig?.tutorEvaluationEnabled,
    tutorEvaluationCategories,
    setTutorEvaluationCategories,
  ])

  return <>{isError ? <ErrorPage onRetry={refetch} /> : isPending ? <LoadingPage /> : children}</>
}
