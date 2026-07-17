import { deleteCoursePhase } from '@core/network/mutations/deleteCoursePhase'
import { postNewCoursePhase } from '@core/network/mutations/postNewCoursePhase'
import { updateCoursePhase } from '@core/network/mutations/updateCoursePhase'
import { updateParticipationDataGraph } from '@core/network/mutations/updateParticipationDataGraph'
import { updatePhaseDataGraph } from '@core/network/mutations/updatePhaseDataGraph'
import { updatePhaseGraph } from '@core/network/mutations/updatePhaseGraph'
import { useMutation } from '@tanstack/react-query'
import type { CreateCoursePhase, UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import type { MetaDataGraphItem } from '../interfaces/courseMetaGraphItem'
import type { CoursePhaseGraphUpdate } from '../interfaces/coursePhaseGraphUpdate'

export function useMutations() {
  const { courseId } = useParams<{ courseId: string }>()
  const { mutateAsync: mutateAsyncPhases, isError: isPhaseError } = useMutation({
    mutationFn: (coursePhase: CreateCoursePhase) => postNewCoursePhase(coursePhase),
  })

  const { mutateAsync: mutateCoursePhaseGraph, isError: isGraphError } = useMutation({
    mutationFn: (update: CoursePhaseGraphUpdate) => updatePhaseGraph(courseId ?? '', update),
  })

  const { mutateAsync: mutateDeletePhase, isError: isDeleteError } = useMutation({
    mutationFn: (coursePhaseId: string) => deleteCoursePhase(coursePhaseId),
  })

  const { mutateAsync: mutateRenamePhase, isError: isRenameError } = useMutation({
    mutationFn: (coursePhase: UpdateCoursePhase) => updateCoursePhase(coursePhase),
  })

  const { mutateAsync: mutatePhaseDataGraph, isError: isPhaseDataGraphError } = useMutation({
    mutationFn: (updatedGraph: MetaDataGraphItem[]) =>
      updatePhaseDataGraph(courseId ?? '', updatedGraph),
  })

  const { mutateAsync: mutateParticipationDataGraph, isError: isParticipationDataGraphError } =
    useMutation({
      mutationFn: (updatedGraph: MetaDataGraphItem[]) =>
        updateParticipationDataGraph(courseId ?? '', updatedGraph),
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
