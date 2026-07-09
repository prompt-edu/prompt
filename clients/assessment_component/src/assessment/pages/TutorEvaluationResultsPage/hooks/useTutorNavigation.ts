import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useGetAllTeams } from '../../hooks/useGetAllTeams'

interface NavigationTutor {
  id: string
  firstName: string
  lastName: string
}

interface TutorNavigation {
  prevTutor?: NavigationTutor
  nextTutor?: NavigationTutor
}

/**
 * Resolves the previous and next tutor relative to the one currently in the
 * URL. Unlike participant navigation, this always cycles through every tutor
 * across all teams.
 */
export const useTutorNavigation = (): TutorNavigation => {
  const { tutorId } = useParams<{ tutorId: string }>()
  const { data: teams } = useGetAllTeams()

  const tutors = useMemo<NavigationTutor[]>(() => {
    const seen = new Set<string>()
    const result: NavigationTutor[] = []
    for (const team of teams) {
      for (const tutor of team.tutors) {
        if (!tutor.id || seen.has(tutor.id)) continue
        seen.add(tutor.id)
        result.push({ id: tutor.id, firstName: tutor.firstName, lastName: tutor.lastName })
      }
    }
    return result
  }, [teams])

  const currentIndex = tutors.findIndex((tutor) => tutor.id === tutorId)

  if (currentIndex === -1 || tutors.length <= 1) {
    return {}
  }

  return {
    prevTutor: tutors[(currentIndex - 1 + tutors.length) % tutors.length],
    nextTutor: tutors[(currentIndex + 1) % tutors.length],
  }
}
