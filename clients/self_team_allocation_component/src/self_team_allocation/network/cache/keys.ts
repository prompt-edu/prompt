type Id = string | undefined

export const selfTeamAllocationKeys = {
  teams: (phaseId: Id) => ['self_team_allocations', phaseId] as const,
  timeframe: (phaseId: Id) => ['timeframe', phaseId] as const,
  // Spelled like Team Allocation's config key, but holds this phase's own config
  config: (phaseId: Id) => ['team_allocation_config', phaseId] as const,
  myParticipation: (phaseId: Id) => ['course_phase_participation', phaseId] as const,
  tutorImport: {
    // Team Allocation reads the same endpoint under the same literal, so the entry is shared
    studentsOfPhase: (phaseId: Id) => ['students', phaseId] as const,
  },
}
