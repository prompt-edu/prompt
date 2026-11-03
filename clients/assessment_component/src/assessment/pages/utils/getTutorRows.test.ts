import type { Team } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'
import { getTeamNamesOfTutor, getTutorRows, getUniqueTutorIds } from './getTutorRows'

const lead = { id: 'lead', firstName: 'Paula', lastName: 'Lead' }
const coach = { id: 'coach', firstName: 'Carla', lastName: 'Coach' }

const teams: Team[] = [
  { id: 'team-a', name: 'Team A', members: [], tutors: [coach, lead] },
  { id: 'team-b', name: 'Team B', members: [], tutors: [lead] },
  { id: 'team-c', name: 'Team C', members: [], tutors: [] },
]

describe('getTutorRows', () => {
  it('lists a tutor once per team they lead, each row with its own id', () => {
    const rows = getTutorRows(teams)

    expect(rows.map((row) => [row.tutorId, row.teamName])).toEqual([
      ['coach', 'Team A'],
      ['lead', 'Team A'],
      ['lead', 'Team B'],
    ])
    expect(new Set(rows.map((row) => row.id)).size).toBe(rows.length)
  })

  it('counts every tutor once', () => {
    expect(getUniqueTutorIds(getTutorRows(teams))).toEqual(['coach', 'lead'])
  })

  it('lists every team a tutor leads', () => {
    expect(getTeamNamesOfTutor(teams, 'lead')).toEqual(['Team A', 'Team B'])
    expect(getTeamNamesOfTutor(teams, 'coach')).toEqual(['Team A'])
    expect(getTeamNamesOfTutor(teams, 'unknown')).toEqual([])
  })
})
