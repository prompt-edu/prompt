import { useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { useTeamStore } from '../../../zustand/useTeamStore'
import { useParticipationStore } from '../../../zustand/useParticipationStore'

interface NavigationMember {
  id: string
  firstName: string
  lastName: string
}

interface ParticipantNavigation {
  prevMember?: NavigationMember
  nextMember?: NavigationMember
}

/**
 * Resolves the previous and next student to assess relative to the participant
 * currently in the URL. Navigation stays within the active team when the
 * student belongs to one, otherwise it cycles through all participations.
 */
export const useParticipantNavigation = (): ParticipantNavigation => {
  const { courseParticipationID } = useParams<{ courseParticipationID: string }>()

  const { teams } = useTeamStore()
  const team = teams.find((t) => t.members.some((member) => member.id === courseParticipationID))
  const { participations } = useParticipationStore()

  const members = useMemo<NavigationMember[]>(() => {
    const teamMembers = team?.members
      .filter((member) => participations.some((p) => p.courseParticipationID === member.id))
      .map((member) => ({
        id: member.id ?? '',
        firstName: member.firstName,
        lastName: member.lastName,
      }))

    if (teamMembers && teamMembers.length > 0) {
      return teamMembers
    }

    return participations.map((p) => ({
      id: p.courseParticipationID,
      firstName: p.student.firstName,
      lastName: p.student.lastName,
    }))
  }, [team?.members, participations])

  const currentIndex = members.findIndex((member) => member.id === courseParticipationID)

  if (currentIndex === -1 || members.length <= 1) {
    return {}
  }

  return {
    prevMember: members[(currentIndex - 1 + members.length) % members.length],
    nextMember: members[(currentIndex + 1) % members.length],
  }
}
