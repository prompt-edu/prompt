import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { resolveNavigationOrder } from '../utils/resolveNavigationOrder'
import { useSorting, useSortSearchParam } from './useSorting'

export interface InterviewNavigationState {
  orderedParticipationIds?: string[]
}

interface InterviewNavigation {
  orderedParticipations: CoursePhaseParticipationWithStudent[]
  currentParticipationId?: string
  navigateToParticipation: (participation: CoursePhaseParticipationWithStudent) => void
}

/**
 * Resolves the order to step through interviewees from a profile. Follows the order the overview
 * showed when the profile was opened, so changing a score or status here does not reshuffle the
 * list being worked through. Without that snapshot (e.g. after a reload) the order is sorted by
 * the `sorting` query param instead.
 */
export const useInterviewNavigation = (): InterviewNavigation => {
  const { studentId } = useParams<{ studentId: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const snapshotIds = (location.state as InterviewNavigationState | null)?.orderedParticipationIds
  const [sortBy] = useSortSearchParam()
  const liveOrder = useSorting(sortBy)

  const currentParticipationId = liveOrder.find(
    (p) => p.student.id === studentId,
  )?.courseParticipationID
  const orderedParticipations = useMemo(() => {
    const participationById = new Map(liveOrder.map((p) => [p.courseParticipationID, p]))
    return resolveNavigationOrder(
      snapshotIds,
      liveOrder.map((p) => p.courseParticipationID),
      currentParticipationId,
    ).flatMap((id) => participationById.get(id) ?? [])
  }, [snapshotIds, liveOrder, currentParticipationId])

  const navigateToParticipation = (participation: CoursePhaseParticipationWithStudent) => {
    navigate(
      { pathname: `../${participation.student.id}`, search: location.search },
      {
        relative: 'path',
        state: {
          orderedParticipationIds: orderedParticipations.map((p) => p.courseParticipationID),
        } satisfies InterviewNavigationState,
      },
    )
  }

  return { orderedParticipations, currentParticipationId, navigateToParticipation }
}
