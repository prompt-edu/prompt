import type { Team } from '@tumaet/prompt-shared-state'
import { getStudentName } from '@tumaet/prompt-ui-components'

export function getTeamMemberName(teams: Team[], courseParticipationID: string): string {
  for (const team of teams) {
    const member = team.members.find((m) => m.id === courseParticipationID)
    if (member) {
      return getStudentName(member)
    }
  }
  return 'Unknown member'
}
