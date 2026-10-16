import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import type { ProfileDetailLocationState } from '../pages/ProfileDetail/ProfileDetailPage'
import { resolveNavigationOrder } from '../utils/resolveNavigationOrder'
import { useParticipationStore } from '../zustand/useParticipationStore'

interface InterviewNavigation {
  orderedParticipations: CoursePhaseParticipationWithStudent[]
  currentParticipationId?: string
  navigateToParticipation: (participation: CoursePhaseParticipationWithStudent) => void
}

/**
 * Resolves the order to step through interviewees from a profile. Follows the order the
 * participants table showed when the profile was opened, including its search and filters, so
 * changing a score or status here does not reshuffle the list being worked through. A profile
 * opened without that order, e.g. in a new tab, steps through all participants instead.
 */
export const useInterviewNavigation = (): InterviewNavigation => {
  const { studentId } = useParams<{ studentId: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const navigationState = (location.state as ProfileDetailLocationState | null) ?? null
  const { participations } = useParticipationStore()

  const currentParticipationId = participations.find(
    (p) => p.student.id === studentId,
  )?.courseParticipationID
  const orderedParticipations = useMemo(() => {
    const participationById = new Map(participations.map((p) => [p.courseParticipationID, p]))
    return resolveNavigationOrder(
      navigationState?.orderedParticipationIds,
      participations.map((p) => p.courseParticipationID),
      currentParticipationId,
    ).flatMap((id) => participationById.get(id) ?? [])
  }, [navigationState?.orderedParticipationIds, participations, currentParticipationId])

  const navigateToParticipation = (participation: CoursePhaseParticipationWithStudent) => {
    navigate(`../${participation.student.id}`, {
      relative: 'path',
      state: {
        tableSearch: navigationState?.tableSearch,
        orderedParticipationIds: orderedParticipations.map((p) => p.courseParticipationID),
      } satisfies ProfileDetailLocationState,
    })
  }

  return { orderedParticipations, currentParticipationId, navigateToParticipation }
}
