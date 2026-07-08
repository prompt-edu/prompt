import type { Team } from '@tumaet/prompt-shared-state'

export function getTeamMemberName(teams: Team[], courseParticipationID: string): string {
  for (const team of teams) {
    const member = team.members.find((m) => m.id === courseParticipationID)
    if (member) {
      return `${member.firstName} ${member.lastName}`
    }
  }
  return 'Unknown member'
}
