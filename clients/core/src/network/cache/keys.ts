type Id = string | undefined

/**
 * Every react-query cache key core reads or writes.
 *
 * The prefix hierarchy is deliberate: `invalidateQueries` matches by prefix, so an unscoped key
 * reaches all of its descendants. Where two entries look like siblings but are not, the comment
 * says so.
 */
export const coreKeys = {
  courses: {
    all: () => ['courses'] as const,
    own: () => ['own_courses'] as const,
    templates: () => ['template-courses'] as const,
    copyability: (courseId: Id) => ['course-copyability', courseId] as const,
    staff: (courseId: Id) => ['courseStaff', courseId] as const,
    myParticipation: (courseId: Id) => ['course_participation', courseId] as const,
  },

  coursePhases: {
    // Owned by @tumaet/prompt-shared-state: react-query is a Module Federation singleton, so this
    // entry is shared with every remote and must keep its literal
    byId: (phaseId: Id) => ['course_phase', phaseId] as const,
    // The archive dialog caches the same phase under its own literal, so it shares neither the
    // entry above nor its invalidations
    byIdInArchiveDialog: (phaseId: Id) => ['coursePhase', phaseId] as const,
    types: () => ['course_phase_types'] as const,
    typesForScope: (scope: 'self' | 'all') => ['coursePhaseType', scope] as const,
  },

  courseGraphs: {
    phase: (courseId: Id) => ['course_phases', 'course_phase_graph', courseId] as const,
    participationData: (courseId: Id) =>
      ['course_phases', 'participation_phase_graph', courseId] as const,
    phaseData: (courseId: Id) => ['course_phases', 'phase_phase_graph', courseId] as const,
  },

  applications: {
    all: () => ['application'] as const,
    // `inPhase` and `ofParticipation` are the same two-element shape holding different ids: the
    // student-facing page caches its own application by phase, the staff dialog caches one
    // application by participation
    inPhase: (phaseId: Id) => ['application', phaseId] as const,
    ofParticipation: (courseParticipationId: Id) => ['application', courseParticipationId] as const,
    ofParticipant: (phaseId: Id, courseParticipationId: Id) =>
      ['application', phaseId, courseParticipationId] as const,
    // Holds the phase's additional score names, not participations, despite sharing their prefix
    additionalScores: (phaseId: Id) => ['application_participations', phaseId] as const,
    participations: {
      all: () => ['application_participations'] as const,
      // The `students` segment does not mean students: this is the phase's participation list
      inPhase: (phaseId: Id) => ['application_participations', 'students', phaseId] as const,
    },
    form: (phaseId: Id) => ['application_form', phaseId] as const,
    exportedAnswers: (phaseId: Id) => ['application_exported_answers', phaseId] as const,
    universityUsers: (searchString: string, phaseId: Id) =>
      ['university_users', searchString, phaseId] as const,
  },

  // The public application pages, which read a different endpoint than `applications.form`
  apply: {
    open: () => ['open_applications'] as const,
    form: (phaseId: Id) => ['applicationForm', phaseId] as const,
  },

  mailCampaigns: {
    inCourse: (courseId: Id) => ['mailCampaigns', courseId] as const,
    byId: (courseId: Id, campaignId: Id) => ['mailCampaign', courseId, campaignId] as const,
    recipientPreview: (courseId: Id, campaignId: Id) =>
      ['mailCampaignRecipientPreview', courseId, campaignId] as const,
  },

  keycloak: {
    status: () => ['keycloakStatus'] as const,
    userSearch: {
      all: () => ['keycloakUserSearch'] as const,
      forQuery: (searchString: string) => ['keycloakUserSearch', searchString] as const,
    },
  },

  instructorNotes: {
    ofStudent: (studentId: Id) => ['instructorNotes', studentId] as const,
    tags: () => ['noteTags'] as const,
  },

  students: {
    byId: (studentId: Id) => ['student', studentId] as const,
    enrollments: (studentId: Id) => ['studentEnrollments', studentId] as const,
  },

  privacy: {
    latest: (resource: string) => ['privacy', `${resource}-latest`] as const,
    create: (resource: string) => ['privacy', `${resource}-create`] as const,
    status: (resource: string, requestId: Id) =>
      ['privacy', `${resource}-status`, requestId] as const,
    admin: {
      exports: () => ['privacy', 'admin', 'exports'] as const,
      deletions: () => ['privacy', 'admin', 'deletions'] as const,
    },
  },

  serviceInfo: {
    core: () => ['serviceInfo-core'] as const,
    // The id sits inside the literal rather than in a second element, so these are not one prefix
    // family and `serviceInfo.core` is not their parent
    ofService: (serviceId: string) => [`serviceInfo-${serviceId}`] as const,
  },

  auditLog: {
    inCourse: (courseId: Id, filters: unknown, limit: number, cursor: unknown) =>
      ['auditLog', courseId, filters, limit, cursor] as const,
    global: (filters: unknown, limit: number, cursor: unknown) =>
      ['auditLog', 'global', filters, limit, cursor] as const,
    status: () => ['auditLogStatus'] as const,
  },

  githubPullRequest: (pullRequestNumber: Id) => ['github-pr', pullRequestNumber] as const,
}
