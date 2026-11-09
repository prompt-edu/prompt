import { type NavigableParticipant, resolveNavigationOrder } from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { useGetAllTeams } from '../../hooks/useGetAllTeams'

export interface TutorDetailLocationState {
  /** The tutor table's order when the results were opened, for previous / next. */
  orderedTutorIds?: string[]
}

interface TutorNavigation {
  orderedTutors: NavigableParticipant[]
  navigateToTutor: (tutor: NavigableParticipant) => void
}

/**
 * Steps through the tutors in the order the tutor table showed them, or through every tutor
 * across all teams when the results were opened without that order, e.g. in a new tab.
 */
export const useTutorNavigation = (): TutorNavigation => {
  const { tutorId } = useParams<{ tutorId: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const snapshotIds = (location.state as TutorDetailLocationState | null)?.orderedTutorIds
  const { data: teams } = useGetAllTeams()

  const orderedTutors = useMemo(() => {
    const tutorById = new Map<string, NavigableParticipant>()
    for (const tutor of teams.flatMap((team) => team.tutors)) {
      if (tutor.id && !tutorById.has(tutor.id)) {
        tutorById.set(tutor.id, { courseParticipationID: tutor.id, student: tutor })
      }
    }
    return resolveNavigationOrder(snapshotIds, [...tutorById.keys()], tutorId).flatMap(
      (id) => tutorById.get(id) ?? [],
    )
  }, [teams, snapshotIds, tutorId])

  const navigateToTutor = (tutor: NavigableParticipant) => {
    navigate(`../${tutor.courseParticipationID}`, {
      relative: 'path',
      state: {
        orderedTutorIds: orderedTutors.map((t) => t.courseParticipationID),
      } satisfies TutorDetailLocationState,
    })
  }

  return { orderedTutors, navigateToTutor }
}
