type Id = string | undefined

export const teamAllocationKeys = {
  teams: (phaseId: Id) => ['team_allocation_team', phaseId] as const,
  skills: (phaseId: Id) => ['team_allocation_skill', phaseId] as const,
  allocations: (phaseId: Id) => ['team_allocations', phaseId] as const,
  config: (phaseId: Id) => ['team_allocation_config', phaseId] as const,
  survey: {
    form: (phaseId: Id) => ['team_allocation_survey_form', phaseId] as const,
    myResponse: (phaseId: Id) => ['team_allocation_student_survey_response', phaseId] as const,
    timeframe: (phaseId: Id) => ['team_allocation_survey_timeframe', phaseId] as const,
    statistics: (phaseId: Id) => ['team_allocation_survey_statistics', phaseId] as const,
  },
  // The TEASE page caches the same teams and skills under its own literals
  tease: {
    teams: (phaseId: Id) => ['tease_teams', phaseId] as const,
    skills: (phaseId: Id) => ['tease_skills', phaseId] as const,
    students: (phaseId: Id) => ['tease_students', phaseId] as const,
  },
  tutorImport: {
    // Self Team Allocation reads the same endpoint under the same literal, so the entry is shared
    studentsOfPhase: (phaseId: Id) => ['students', phaseId] as const,
    studentSearch: (searchString: string) => ['student-search', searchString] as const,
  },
  // Owned by @tumaet/prompt-shared-state: react-query is a Module Federation singleton, so this
  // entry is shared with core and must keep its literal
  participants: (phaseId: Id) => ['participants', phaseId] as const,
}
