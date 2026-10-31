import { useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import type { AssessmentParticipationWithStudent } from '../../interfaces/assessmentParticipationWithStudent'
import { resolveNavigationOrder } from '../utils/resolveNavigationOrder'
import { useGetCoursePhaseParticipations } from './useGetCoursePhaseParticipations'

export interface ParticipantDetailLocationState {
  orderedParticipationIds?: string[]
}

interface ParticipantNavigationOrder {
  orderedParticipations: AssessmentParticipationWithStudent[]
  navigateToParticipation: (participation: AssessmentParticipationWithStudent) => void
}

export const useParticipantNavigation = (): ParticipantNavigationOrder => {
  const { courseParticipationID } = useParams<{ courseParticipationID: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const snapshotIds = (location.state as ParticipantDetailLocationState | null)
    ?.orderedParticipationIds
  const { data: participations } = useGetCoursePhaseParticipations()

  const orderedParticipations = useMemo(() => {
    const participationById = new Map(participations.map((p) => [p.courseParticipationID, p]))
    return resolveNavigationOrder(
      snapshotIds,
      participations.map((p) => p.courseParticipationID),
      courseParticipationID,
    ).flatMap((id) => participationById.get(id) ?? [])
  }, [snapshotIds, participations, courseParticipationID])

  const navigateToParticipation = (participation: AssessmentParticipationWithStudent) => {
    navigate(`../${participation.courseParticipationID}`, {
      relative: 'path',
      state: {
        orderedParticipationIds: orderedParticipations.map((p) => p.courseParticipationID),
      } satisfies ParticipantDetailLocationState,
    })
  }

  return { orderedParticipations, navigateToParticipation }
}
