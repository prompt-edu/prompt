import { QueryClient } from '@tanstack/react-query'
import { beforeEach, describe, expect, it } from 'vitest'

import { coreCache } from './events'
import { coreKeys } from './keys'

const COURSE = 'course-1'
const OTHER_COURSE = 'course-2'
const PHASE = 'phase-1'
const OTHER_PHASE = 'phase-2'
const PARTICIPATION = 'participation-1'
const CAMPAIGN = 'campaign-1'
const STUDENT = 'student-1'

let queryClient: QueryClient

const seed = (...keys: readonly (readonly unknown[])[]): void => {
  for (const key of keys) {
    queryClient.setQueryData(key, 'seeded')
  }
}

const isInvalidated = (key: readonly unknown[]): boolean =>
  queryClient.getQueryState(key)?.isInvalidated === true

beforeEach(() => {
  queryClient = new QueryClient()
})

describe('coursesChanged', () => {
  it('invalidates the course list', () => {
    seed(coreKeys.courses.all())

    coreCache.coursesChanged(queryClient)

    expect(isInvalidated(coreKeys.courses.all())).toBe(true)
  })

  it('leaves the own-course list alone, which is a separate cache', () => {
    seed(coreKeys.courses.own())

    coreCache.coursesChanged(queryClient)

    expect(isInvalidated(coreKeys.courses.own())).toBe(false)
  })
})

describe('courseCreated', () => {
  it('resolves only once the course list has been fetched again', async () => {
    let fetches = 0
    await queryClient.fetchQuery({
      queryKey: coreKeys.courses.all(),
      queryFn: () => {
        fetches += 1
        return 'courses'
      },
    })

    await coreCache.courseCreated(queryClient)

    expect(fetches).toBe(2)
  })
})

describe('courseStaffChanged', () => {
  it('invalidates this course staff and every cached user search', () => {
    seed(
      coreKeys.courses.staff(COURSE),
      coreKeys.keycloak.userSearch.all(),
      coreKeys.keycloak.userSearch.forQuery('ab'),
    )

    coreCache.courseStaffChanged(queryClient, COURSE)

    expect(isInvalidated(coreKeys.courses.staff(COURSE))).toBe(true)
    expect(isInvalidated(coreKeys.keycloak.userSearch.all())).toBe(true)
    expect(isInvalidated(coreKeys.keycloak.userSearch.forQuery('ab'))).toBe(true)
  })

  it('leaves another course staff alone', () => {
    seed(coreKeys.courses.staff(OTHER_COURSE))

    coreCache.courseStaffChanged(queryClient, COURSE)

    expect(isInvalidated(coreKeys.courses.staff(OTHER_COURSE))).toBe(false)
  })
})

describe('coursePhaseChanged', () => {
  it('invalidates the shared course phase entry', () => {
    seed(coreKeys.coursePhases.byId(PHASE))

    coreCache.coursePhaseChanged(queryClient, PHASE)

    expect(isInvalidated(coreKeys.coursePhases.byId(PHASE))).toBe(true)
  })

  it('does not reach the archive dialog copy, which caches under its own literal', () => {
    seed(coreKeys.coursePhases.byIdInArchiveDialog(PHASE))

    coreCache.coursePhaseChanged(queryClient, PHASE)

    expect(isInvalidated(coreKeys.coursePhases.byIdInArchiveDialog(PHASE))).toBe(false)
  })
})

describe('coursePhaseGraphSaved', () => {
  it('reaches none of the graphs, because it sends one composite key rather than four', () => {
    seed(
      coreKeys.courses.all(),
      coreKeys.courseGraphs.phase(COURSE),
      coreKeys.courseGraphs.participationData(COURSE),
      coreKeys.courseGraphs.phaseData(COURSE),
      coreKeys.coursePhases.types(),
    )

    coreCache.coursePhaseGraphSaved(queryClient)

    expect(isInvalidated(coreKeys.courses.all())).toBe(false)
    expect(isInvalidated(coreKeys.courseGraphs.phase(COURSE))).toBe(false)
    expect(isInvalidated(coreKeys.courseGraphs.participationData(COURSE))).toBe(false)
    expect(isInvalidated(coreKeys.courseGraphs.phaseData(COURSE))).toBe(false)
    expect(isInvalidated(coreKeys.coursePhases.types())).toBe(false)
  })
})

describe('applicationsImported', () => {
  it('invalidates both participation caches of this phase and their descendants', () => {
    seed(
      coreKeys.applications.participations.inPhase(PHASE),
      coreKeys.applications.participations.students(PHASE),
    )

    coreCache.applicationsImported(queryClient, PHASE)

    expect(isInvalidated(coreKeys.applications.participations.inPhase(PHASE))).toBe(true)
    expect(isInvalidated(coreKeys.applications.participations.students(PHASE))).toBe(true)
  })

  it('leaves another phase alone', () => {
    seed(coreKeys.applications.participations.students(OTHER_PHASE))

    coreCache.applicationsImported(queryClient, PHASE)

    expect(isInvalidated(coreKeys.applications.participations.students(OTHER_PHASE))).toBe(false)
  })
})

describe('applicationParticipantsChanged', () => {
  it('invalidates only the student rows of this phase', () => {
    seed(
      coreKeys.applications.participations.students(PHASE),
      coreKeys.applications.participations.inPhase(PHASE),
    )

    coreCache.applicationParticipantsChanged(queryClient, PHASE)

    expect(isInvalidated(coreKeys.applications.participations.students(PHASE))).toBe(true)
    expect(isInvalidated(coreKeys.applications.participations.inPhase(PHASE))).toBe(false)
  })
})

describe('applicationAssessmentSaved', () => {
  it('reaches every phase, because the participation key it uses is unscoped', () => {
    seed(
      coreKeys.applications.participations.students(PHASE),
      coreKeys.applications.participations.students(OTHER_PHASE),
    )

    coreCache.applicationAssessmentSaved(queryClient)

    expect(isInvalidated(coreKeys.applications.participations.students(PHASE))).toBe(true)
    expect(isInvalidated(coreKeys.applications.participations.students(OTHER_PHASE))).toBe(true)
  })

  it('leaves the applications themselves alone', () => {
    seed(coreKeys.applications.ofParticipation(PARTICIPATION))

    coreCache.applicationAssessmentSaved(queryClient)

    expect(isInvalidated(coreKeys.applications.ofParticipation(PARTICIPATION))).toBe(false)
  })
})

describe('studentUniversityDataChanged', () => {
  it('invalidates every application and every participation row', () => {
    seed(
      coreKeys.applications.ofParticipation(PARTICIPATION),
      coreKeys.applications.inPhase(PHASE),
      coreKeys.applications.participations.students(OTHER_PHASE),
    )

    coreCache.studentUniversityDataChanged(queryClient)

    expect(isInvalidated(coreKeys.applications.ofParticipation(PARTICIPATION))).toBe(true)
    expect(isInvalidated(coreKeys.applications.inPhase(PHASE))).toBe(true)
    expect(isInvalidated(coreKeys.applications.participations.students(OTHER_PHASE))).toBe(true)
  })

  it('leaves the application form alone, which is a separate cache', () => {
    seed(coreKeys.applications.form(PHASE))

    coreCache.studentUniversityDataChanged(queryClient)

    expect(isInvalidated(coreKeys.applications.form(PHASE))).toBe(false)
  })
})

describe('myApplicationSubmitted', () => {
  it('invalidates the application of this phase only', () => {
    seed(coreKeys.applications.inPhase(PHASE), coreKeys.applications.inPhase(OTHER_PHASE))

    coreCache.myApplicationSubmitted(queryClient, PHASE)

    expect(isInvalidated(coreKeys.applications.inPhase(PHASE))).toBe(true)
    expect(isInvalidated(coreKeys.applications.inPhase(OTHER_PHASE))).toBe(false)
  })
})

describe('mail campaign events', () => {
  it('invalidates the preview on an edit, since targeting may have changed', () => {
    seed(
      coreKeys.mailCampaigns.inCourse(COURSE),
      coreKeys.mailCampaigns.byId(COURSE, CAMPAIGN),
      coreKeys.mailCampaigns.recipientPreview(COURSE, CAMPAIGN),
    )

    coreCache.mailCampaignChanged(queryClient, COURSE, CAMPAIGN)

    expect(isInvalidated(coreKeys.mailCampaigns.inCourse(COURSE))).toBe(true)
    expect(isInvalidated(coreKeys.mailCampaigns.byId(COURSE, CAMPAIGN))).toBe(true)
    expect(isInvalidated(coreKeys.mailCampaigns.recipientPreview(COURSE, CAMPAIGN))).toBe(true)
  })

  it('leaves the preview alone on a send, which cannot change targeting', () => {
    seed(
      coreKeys.mailCampaigns.byId(COURSE, CAMPAIGN),
      coreKeys.mailCampaigns.recipientPreview(COURSE, CAMPAIGN),
    )

    coreCache.mailCampaignSent(queryClient, COURSE, CAMPAIGN)

    expect(isInvalidated(coreKeys.mailCampaigns.byId(COURSE, CAMPAIGN))).toBe(true)
    expect(isInvalidated(coreKeys.mailCampaigns.recipientPreview(COURSE, CAMPAIGN))).toBe(false)
  })

  it('touches only the list when a campaign is added or removed', () => {
    seed(coreKeys.mailCampaigns.inCourse(COURSE), coreKeys.mailCampaigns.byId(COURSE, CAMPAIGN))

    coreCache.mailCampaignListChanged(queryClient, COURSE)

    expect(isInvalidated(coreKeys.mailCampaigns.inCourse(COURSE))).toBe(true)
    expect(isInvalidated(coreKeys.mailCampaigns.byId(COURSE, CAMPAIGN))).toBe(false)
  })
})

describe('instructor note events', () => {
  it('invalidates the tag list without touching the notes', () => {
    seed(coreKeys.instructorNotes.tags(), coreKeys.instructorNotes.ofStudent(STUDENT))

    coreCache.noteTagsChanged(queryClient)

    expect(isInvalidated(coreKeys.instructorNotes.tags())).toBe(true)
    expect(isInvalidated(coreKeys.instructorNotes.ofStudent(STUDENT))).toBe(false)
  })

  it('invalidates the notes of one student only', () => {
    seed(coreKeys.instructorNotes.ofStudent(STUDENT), coreKeys.instructorNotes.ofStudent('other'))

    coreCache.instructorNotesChanged(queryClient, STUDENT)

    expect(isInvalidated(coreKeys.instructorNotes.ofStudent(STUDENT))).toBe(true)
    expect(isInvalidated(coreKeys.instructorNotes.ofStudent('other'))).toBe(false)
  })
})

describe('privacy events', () => {
  it('keeps the admin deletion and export lists apart', () => {
    seed(coreKeys.privacy.admin.deletions(), coreKeys.privacy.admin.exports())

    coreCache.privacyDeletionsChanged(queryClient)

    expect(isInvalidated(coreKeys.privacy.admin.deletions())).toBe(true)
    expect(isInvalidated(coreKeys.privacy.admin.exports())).toBe(false)

    coreCache.privacyExportsChanged(queryClient)

    expect(isInvalidated(coreKeys.privacy.admin.exports())).toBe(true)
  })
})
