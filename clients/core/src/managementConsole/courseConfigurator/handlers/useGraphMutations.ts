import { coreApi } from '@core/network/api'
import { useMutation } from '@tanstack/react-query'
import type { CreateCoursePhase, UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import type { MetaDataGraphItem } from '../interfaces/courseMetaGraphItem'
import type { CoursePhaseGraphUpdate } from '../interfaces/coursePhaseGraphUpdate'

export function useMutations() {
  const { courseId } = useParams<{ courseId: string }>()
  const { mutateAsync: mutateAsyncPhases, isError: isPhaseError } = useMutation({
    mutationFn: (coursePhase: CreateCoursePhase) => coreApi.coursePhases.create(coursePhase),
  })

  const { mutateAsync: mutateCoursePhaseGraph, isError: isGraphError } = useMutation({
    mutationFn: (update: CoursePhaseGraphUpdate) =>
      coreApi.courseGraphs.savePhase(courseId ?? '', update),
  })

  const { mutateAsync: mutateDeletePhase, isError: isDeleteError } = useMutation({
    mutationFn: (coursePhaseId: string) => coreApi.coursePhases.remove(coursePhaseId),
  })

  const { mutateAsync: mutateRenamePhase, isError: isRenameError } = useMutation({
    mutationFn: (coursePhase: UpdateCoursePhase) => coreApi.coursePhases.update(coursePhase),
  })

  const { mutateAsync: mutatePhaseDataGraph, isError: isPhaseDataGraphError } = useMutation({
    mutationFn: (updatedGraph: MetaDataGraphItem[]) =>
      coreApi.courseGraphs.savePhaseData(courseId ?? '', updatedGraph),
  })

  const { mutateAsync: mutateParticipationDataGraph, isError: isParticipationDataGraphError } =
    useMutation({
      mutationFn: (updatedGraph: MetaDataGraphItem[]) =>
        coreApi.courseGraphs.saveParticipationData(courseId ?? '', updatedGraph),
      onSuccess: () => {
        // this is the last executed mutation and on this we want to reload!
        window.location.reload()
      },
    })

  const isMutationError =
    isPhaseError ||
    isGraphError ||
    isPhaseDataGraphError ||
    isParticipationDataGraphError ||
    isDeleteError ||
    isRenameError

  return {
    mutateAsyncPhases,
    mutateCoursePhaseGraph,
    mutateParticipationDataGraph,
    mutatePhaseDataGraph,
    mutateDeletePhase,
    mutateRenamePhase,
    isMutationError,
  }
}
