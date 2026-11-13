import type { QueryClient } from '@tanstack/react-query'

import { coreKeys } from './keys'

type Id = string | undefined
type CacheKeys = readonly (readonly unknown[])[]

/**
 * The part of a key that is safe to invalidate.
 *
 * Keys are built from route params, so a scoping id can be `undefined`. `invalidateQueries` matches
 * by prefix and compares every segment, so a key holding `undefined` matches nothing cached and the
 * invalidation silently reaches no cache at all. Truncating at the missing segment over-invalidates
 * instead, which costs a refetch, where the full key would leave stale data on screen. Every key in
 * `coreKeys` opens with a literal, so the result is never the whole cache.
 */
const definedPrefixOf = (queryKey: readonly unknown[]): readonly unknown[] => {
  const missing = queryKey.indexOf(undefined)
  return missing === -1 ? queryKey : queryKey.slice(0, missing)
}

const invalidate = (queryClient: QueryClient, keys: CacheKeys): void => {
  for (const queryKey of keys) {
    queryClient.invalidateQueries({ queryKey: definedPrefixOf(queryKey) })
  }
}

/**
 * Drops a query from the cache once no mounted component observes it.
 *
 * A query removed while a component still observes it is rebuilt empty on that component's next
 * render and fetched again. The page that deleted an entity still renders until the navigation away
 * commits, so removing the entity's query right away would refetch it into a 404.
 */
const removeOnceUnobserved = (queryClient: QueryClient, queryKey: readonly unknown[]): void => {
  const queryCache = queryClient.getQueryCache()
  const query = queryCache.find({ queryKey, exact: true })
  if (!query) {
    return
  }
  if (query.getObserversCount() === 0) {
    queryCache.remove(query)
    return
  }
  const unsubscribe = queryCache.subscribe((event) => {
    if (event.query !== query) {
      return
    }
    if (event.type === 'removed') {
      unsubscribe()
    } else if (event.type === 'observerRemoved' && query.getObserversCount() === 0) {
      unsubscribe()
      queryCache.remove(query)
    }
  })
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

  applicationAssessmentSaved: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [coreKeys.applications.participations.inPhase(phaseId)]),

  // The whole student record is written, and the student's rows render in the applications and
  // participations of every phase, which the call site does not know
  studentUniversityDataChanged: (queryClient: QueryClient, studentId: Id): void =>
    invalidate(queryClient, [
      coreKeys.students.byId(studentId),
      coreKeys.applications.all(),
      coreKeys.applications.participations.all(),
      // Unscoped, so every cached search re-runs: the write sets the fields they search
      coreKeys.applications.universityUsers.all(),
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

  announcementsChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.announcements.all()]),

  noteTagsChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.instructorNotes.tags()]),

  instructorNotesChanged: (queryClient: QueryClient, studentId: Id): void =>
    invalidate(queryClient, [coreKeys.instructorNotes.ofStudent(studentId)]),

  studyProgramsChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.studyPrograms.all()]),

  studyProgramRenamed: (queryClient: QueryClient): void =>
    invalidate(queryClient, [
      coreKeys.studyPrograms.all(),
      coreKeys.students.all(),
      coreKeys.applications.all(),
      coreKeys.applications.participations.all(),
    ]),

  orgsChanged: (queryClient: QueryClient): void => invalidate(queryClient, [coreKeys.orgs.all()]),

  // The deleted org is removed rather than invalidated: a refetch would only 404, and a cached copy
  // would render it again on browser Back until that refetch returns
  orgDeleted: (queryClient: QueryClient, orgId: string): void => {
    queryClient.invalidateQueries({ queryKey: coreKeys.orgs.all(), exact: true })
    queryClient.invalidateQueries({ queryKey: coreKeys.orgs.own(), exact: true })
    removeOnceUnobserved(queryClient, coreKeys.orgs.byId(orgId))
    removeOnceUnobserved(queryClient, coreKeys.orgs.members(orgId))
  },

  orgMembersChanged: (queryClient: QueryClient, orgId: string): void =>
    invalidate(queryClient, [coreKeys.orgs.members(orgId)]),

  privacyDeletionsChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.privacy.admin.deletions()]),

  privacyExportsChanged: (queryClient: QueryClient): void =>
    invalidate(queryClient, [coreKeys.privacy.admin.exports()]),
}
