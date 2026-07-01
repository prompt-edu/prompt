// src/hooks/useCourseSetup.ts

import { getAdditionalScoreNames } from '@core/network/queries/additionalScoreNames'
import { getApplicationForm } from '@core/network/queries/applicationForm'
import { getParticipationDataGraph } from '@core/network/queries/courseParticipationDataGraph'
import { getPhaseDataGraph } from '@core/network/queries/coursePhaseDataGraph'
import { getCoursePhaseGraph } from '@core/network/queries/coursePhaseGraph'
import { getAllCoursePhaseTypes } from '@core/network/queries/coursePhaseTypes'
import { useQuery } from '@tanstack/react-query'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { AdditionalScore } from '../../applicationAdministration/interfaces/additionalScore/additionalScore'
import { ApplicationForm } from '../../applicationAdministration/interfaces/form/applicationForm'
import { MetaDataGraphItem } from '../interfaces/courseMetaGraphItem'
import { CoursePhaseGraphItem } from '../interfaces/coursePhaseGraphItem'
import { CoursePhaseType } from '../interfaces/coursePhaseType'
import { useCourseConfigurationState } from '../zustand/useCourseConfigurationStore'

export function useCourseConfiguratorDataSetup() {
  const { courses } = useCourseStore()
  const courseId = useParams<{ courseId: string }>().courseId || ''
  const course = courses.find((c) => c.id === courseId)
  const {
    setCoursePhaseTypes,
    appendCoursePhaseType,
    setCoursePhaseGraph,
    setCoursePhases,
    setParticipationDataGraph,
    setPhaseDataGraph,
  } = useCourseConfigurationState()

  // Flags to delay canvas load until everything is ready
  const [finishedGraphSetup, setFinishedGraphSetup] = useState<string>('')
  const [finishedCoursePhaseSetup, setFinishedCoursePhaseSetup] = useState<string>('')

  // Queries
  const {
    data: fetchedCoursePhaseTypes,
    isPending: isCoursePhaseTypesPending,
    error,
    isError: isCoursePhaseTypesError,
    refetch: refetchCoursePhaseTypes,
  } = useQuery<CoursePhaseType[]>({
    queryKey: ['course_phase_types'],
    queryFn: getAllCoursePhaseTypes,
  })

  const {
    data: fetchedCourseGraph,
    isPending: isGraphPending,
    error: graphError,
    isError: isGraphError,
    refetch: refetchGraph,
  } = useQuery<CoursePhaseGraphItem[]>({
    queryKey: ['course_phases', 'course_phase_graph', courseId],
    queryFn: () => getCoursePhaseGraph(courseId),
    enabled: !!courseId,
  })

  const {
    data: fetchedParticipationDataGraph,
    isPending: isParticipationGraphPending,
    error: participationGraphError,
    isError: isParticipationGraphError,
    refetch: refetchParticipationGraph,
  } = useQuery<MetaDataGraphItem[]>({
    queryKey: ['course_phases', 'participation_phase_graph', courseId],
    queryFn: () => getParticipationDataGraph(courseId),
    enabled: !!courseId,
  })

  const {
    data: fetchedPhaseDataGraph,
    isPending: isPhaseGraphPending,
    error: phaseGraphError,
    isError: isPhaseGraphError,
    refetch: refetchPhaseGraph,
  } = useQuery<MetaDataGraphItem[]>({
    queryKey: ['course_phases', 'phase_phase_graph', courseId],
    queryFn: () => getPhaseDataGraph(courseId),
    enabled: !!courseId,
  })

  // Get the application phase from the course phases.
  const applicationPhase = course?.coursePhases.find(
    (phase) => phase.coursePhaseType === 'Application',
  )

  const {
    data: fetchedApplicationForm,
    isPending: isFetchingApplicationForm,
    isError: isApplicationFormError,
  } = useQuery<ApplicationForm>({
    queryKey: ['application_form', applicationPhase?.id],
    queryFn: () => getApplicationForm(applicationPhase?.id || ''),
    enabled: !!applicationPhase?.id,
  })

  const {
    data: fetchedAdditionalScores,
    isPending: isAdditionalScoresPending,
    isError: isAdditionalScoresError,
  } = useQuery<AdditionalScore[]>({
    queryKey: ['application_participations', applicationPhase?.id],
    queryFn: () => getAdditionalScoreNames(applicationPhase?.id || ''),
    enabled: !!applicationPhase?.id,
  })

  // Combine error and pending flags.
  const isError =
    isCoursePhaseTypesError ||
    isGraphError ||
    isParticipationGraphError ||
    isPhaseGraphError ||
    isApplicationFormError ||
    isAdditionalScoresError

  const isPending =
    isCoursePhaseTypesPending ||
    isGraphPending ||
    isParticipationGraphPending ||
    isPhaseGraphPending ||
    (!!applicationPhase?.id && (isFetchingApplicationForm || isAdditionalScoresPending))

  // Set up course phase types with additional metadata for application phase.
  useEffect(() => {
    if (fetchedCoursePhaseTypes) {
      setCoursePhaseTypes(fetchedCoursePhaseTypes) // Clear existing state
      // TODO: maybe re-incorporate the exported application answers
    }
  }, [
    fetchedCoursePhaseTypes,
    fetchedApplicationForm,
    fetchedAdditionalScores,
    isAdditionalScoresPending,
    appendCoursePhaseType,
    setCoursePhaseTypes,
  ])

  // Set up the course graph and metadata graph.
  useEffect(() => {
    if (fetchedCourseGraph && fetchedParticipationDataGraph && fetchedPhaseDataGraph) {
      setCoursePhaseGraph([...fetchedCourseGraph])
      setParticipationDataGraph([...fetchedParticipationDataGraph])
      setPhaseDataGraph([...fetchedPhaseDataGraph])
      setFinishedGraphSetup(courseId)
    }
  }, [
    fetchedCourseGraph,
    fetchedParticipationDataGraph,
    setCoursePhaseGraph,
    courseId,
    setParticipationDataGraph,
    setPhaseDataGraph,
    fetchedPhaseDataGraph,
  ])

  // Set up the course phases.
  useEffect(() => {
    if (course) {
      setCoursePhases(
        course.coursePhases.map((phase) => ({
          ...phase,
          position: { x: 0, y: 0 },
          restrictedMetaData: [],
          studentReadableData: [],
        })),
      )
    } else {
      console.error('Course not found')
    }
    setFinishedCoursePhaseSetup(courseId)
  }, [course, setCoursePhases, courseId])

  // A flag to indicate that both graph and course phase setups are finished.
  const finishedSetup = finishedGraphSetup === courseId && finishedCoursePhaseSetup === courseId

  // A combined refetch function.
  const refetchAll = () => {
    refetchCoursePhaseTypes()
    refetchGraph()
    refetchParticipationGraph()
    refetchPhaseGraph()
  }

  return {
    courseId,
    isError,
    isPending,
    error: error || graphError || participationGraphError || phaseGraphError,
    finishedSetup,
    refetchAll,
  }
}
