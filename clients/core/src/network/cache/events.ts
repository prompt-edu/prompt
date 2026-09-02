import type { QueryClient } from '@tanstack/react-query'

import { coreKeys } from './keys'

type Id = string | undefined
type CacheKeys = readonly (readonly unknown[])[]

const invalidate = (queryClient: QueryClient, keys: CacheKeys): void => {
  for (const queryKey of keys) {
    queryClient.invalidateQueries({ queryKey })
  }
}

/**
 * The writes core performs, and the caches each one makes stale.
 *
 * Hooks name the event; the fan-out lives here, so a new cache that depends on an existing write is
 * one edit rather than one per call site.
 */
export const coreCache = {
  coursesChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.courses.all()]),

  // The caller navigates into the new course as soon as this resolves, so the list has to be back
  // before then: an invalidation alone would only mark it stale.
  courseCreated: async (queryClient: QueryClient): Promise<void> => {
    await queryClient.invalidateQueries({ queryKey: coreKeys.courses.all() })
    await queryClient.refetchQueries({ queryKey: coreKeys.courses.all() })
  },

  courseCopyabilityChanged: (queryClient: QueryClient, courseId: Id): void =>
    invalidate(queryClient, [coreKeys.courses.copyability(courseId)]),

  courseStaffChanged: (queryClient: QueryClient, courseId: Id): void =>
    invalidate(queryClient, [
      coreKeys.courses.staff(courseId),
      // Unscoped, so every cached search re-runs: membership decides what the results may offer
      coreKeys.keycloak.userSearch.all(),
    ]),

  coursePhaseChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [coreKeys.coursePhases.byId(phaseId)]),

  // One composite key rather than one key per graph, which is what the configurator has always
  // sent. `invalidateQueries` matches by prefix, so nothing caches under it and nothing is
  // invalidated; the configurator reloads its graphs by other means.
  coursePhaseGraphSaved: (queryClient: QueryClient): void =>
    invalidate(queryClient, [
      [
        'courses',
        'participation_data_phase_graph',
        'phase_data_phase_graph',
        'course_phase_types',
        'course_phase_graph',
      ],
    ]),

  applicationsImported: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      coreKeys.applications.additionalScores(phaseId),
      coreKeys.applications.participations.inPhase(phaseId),
    ]),

  additionalScoresUploaded: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      coreKeys.applications.additionalScores(phaseId),
      coreKeys.applications.participations.inPhase(phaseId),
    ]),

  applicationParticipantsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [coreKeys.applications.participations.inPhase(phaseId)]),

  // Unscoped, so it reaches every phase rather than the one that was assessed
  applicationAssessmentSaved: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.applications.participations.all()]),

  // A student gaining a university account changes the applications that name them and the
  // participation rows that render them, in every phase
  studentUniversityDataChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [
      coreKeys.applications.all(),
      coreKeys.applications.participations.all(),
    ]),

  applicationFormChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [coreKeys.applications.form(phaseId)]),

  myApplicationSubmitted: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [coreKeys.applications.inPhase(phaseId)]),

  mailCampaignListChanged: (queryClient: QueryClient, courseId: Id): void =>
    invalidate(queryClient, [coreKeys.mailCampaigns.inCourse(courseId)]),

  mailCampaignChanged: (queryClient: QueryClient, courseId: Id, campaignId: Id): void =>
    invalidate(queryClient, [
      coreKeys.mailCampaigns.inCourse(courseId),
      coreKeys.mailCampaigns.byId(courseId, campaignId),
      // Targeting (phase/statuses) may have changed, so the cached preview is stale
      coreKeys.mailCampaigns.recipientPreview(courseId, campaignId),
    ]),

  mailCampaignSent: (queryClient: QueryClient, courseId: Id, campaignId: Id): void =>
    invalidate(queryClient, [
      coreKeys.mailCampaigns.inCourse(courseId),
      coreKeys.mailCampaigns.byId(courseId, campaignId),
    ]),

  noteTagsChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.instructorNotes.tags()]),

  instructorNotesChanged: (queryClient: QueryClient, studentId: Id): void =>
    invalidate(queryClient, [coreKeys.instructorNotes.ofStudent(studentId)]),

  privacyDeletionsChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.privacy.admin.deletions()]),

  privacyExportsChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.privacy.admin.exports()]),
}
