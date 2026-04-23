import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { TriangleAlert } from 'lucide-react'

import { useCourseStore, CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  ErrorPage,
  LoadingPage,
} from '@tumaet/prompt-ui-components'
import { getOwnCoursePhaseParticipation } from '@tumaet/prompt-shared-state'

import { useCoursePhaseConfigStore } from '../zustand/useCoursePhaseConfigStore'
import { useTeamStore } from '../zustand/useTeamStore'
import { useMyParticipationStore } from '../zustand/useMyParticipationStore'
import { useSelfEvaluationCategoryStore } from '../zustand/useSelfEvaluationCategoryStore'
import { usePeerEvaluationCategoryStore } from '../zustand/usePeerEvaluationCategoryStore'
import { useTutorEvaluationCategoryStore } from '../zustand/useTutorEvaluationCategoryStore'
import { useEvaluationStore } from '../zustand/useEvaluationStore'

import { useGetCoursePhaseConfig } from './hooks/useGetCoursePhaseConfig'
import { useGetAllTeams } from './hooks/useGetAllTeams'
import { useGetSelfEvaluationCategoriesWithCompetencies } from './hooks/useGetSelfEvaluationCategoriesWithCompetencies'
import { useGetPeerEvaluationCategoriesWithCompetencies } from './hooks/useGetPeerEvaluationCategoriesWithCompetencies'
import { useGetTutorEvaluationCategoriesWithCompetencies } from './hooks/useGetTutorEvaluationCategoriesWithCompetencies'
import { useGetMyEvaluationCompletions } from './hooks/useGetMyEvaluationCompletions'
import { useGetMyEvaluations } from './hooks/useGetMyEvaluations'

interface EvaluationDataShellProps {
  children: React.ReactNode
}

export const EvaluationDataShell = ({ children }: EvaluationDataShellProps) => {
  const { isStudentOfCourse } = useCourseStore()
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const { setCoursePhaseConfig } = useCoursePhaseConfigStore()
  const { setSelfEvaluationCategories } = useSelfEvaluationCategoryStore()
  const { setPeerEvaluationCategories } = usePeerEvaluationCategoryStore()
  const { setTutorEvaluationCategories } = useTutorEvaluationCategoryStore()
  const { setTeams } = useTeamStore()
  const { setMyParticipation } = useMyParticipationStore()
  const {
    setSelfEvaluationCompletion,
    setPeerEvaluationCompletions,
    setTutorEvaluationCompletions,
    setSelfEvaluations,
    setPeerEvaluations,
    setTutorEvaluations,
  } = useEvaluationStore()

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

  const {
    data: teams,
    isPending: isTeamsPending,
    isError: isTeamsError,
    refetch: refetchTeams,
  } = useGetAllTeams()

  const {
    data: participation,
    isPending: isParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useQuery<CoursePhaseParticipationWithStudent>({
    queryKey: ['course_phase_participation', phaseId],
    queryFn: () => getOwnCoursePhaseParticipation(phaseId ?? ''),
    enabled: isStudent,
  })

  const {
    selfEvaluationCompletion: selfEvaluationCompletion,
    peerEvaluationCompletions: peerEvaluationCompletions,
    tutorEvaluationCompletions: tutorEvaluationCompletions,
    isPending: isCompletionPending,
    isError: isCompletionError,
    refetch: refetchCompletion,
  } = useGetMyEvaluationCompletions({ enabled: isStudent })

  const {
    selfEvaluations: selfEvaluations,
    peerEvaluations: peerEvaluations,
    tutorEvaluations: tutorEvaluations,
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

  useEffect(() => {
    if (coursePhaseConfig) {
      setCoursePhaseConfig(coursePhaseConfig)
    }
  }, [coursePhaseConfig, setCoursePhaseConfig])

  useEffect(() => {
    if (teams) {
      setTeams(teams)
    }
  }, [teams, setTeams])

  useEffect(() => {
    if (participation && isStudent) {
      setMyParticipation(participation)
    }
  }, [participation, setMyParticipation, isStudent])

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

  useEffect(() => {
    if (selfEvaluationCompletion && isStudent) {
      setSelfEvaluationCompletion(selfEvaluationCompletion)
    }
  }, [selfEvaluationCompletion, setSelfEvaluationCompletion, isStudent])

  useEffect(() => {
    if (peerEvaluationCompletions && isStudent) {
      setPeerEvaluationCompletions(peerEvaluationCompletions)
    }
  }, [peerEvaluationCompletions, setPeerEvaluationCompletions, isStudent])

  useEffect(() => {
    if (tutorEvaluationCompletions && isStudent) {
      setTutorEvaluationCompletions(tutorEvaluationCompletions)
    }
  }, [tutorEvaluationCompletions, setTutorEvaluationCompletions, isStudent])

  useEffect(() => {
    if (selfEvaluations && isStudent) {
      setSelfEvaluations(selfEvaluations)
    }
  }, [selfEvaluations, setSelfEvaluations, isStudent])

  useEffect(() => {
    if (peerEvaluations && isStudent) {
      setPeerEvaluations(peerEvaluations)
    }
  }, [peerEvaluations, setPeerEvaluations, isStudent])

  useEffect(() => {
    if (tutorEvaluations && isStudent) {
      setTutorEvaluations(tutorEvaluations)
    }
  }, [tutorEvaluations, setTutorEvaluations, isStudent])

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
