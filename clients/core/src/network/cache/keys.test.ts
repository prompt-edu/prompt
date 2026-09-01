import { describe, expect, it } from 'vitest'

import { coreKeys } from './keys'

const COURSE = 'course-1'
const PHASE = 'phase-1'
const PARTICIPATION = 'participation-1'
const CAMPAIGN = 'campaign-1'
const STUDENT = 'student-1'

describe('coreKeys', () => {
  it('builds the course keys, each under its own literal', () => {
    expect(coreKeys.courses.all()).toEqual(['courses'])
    expect(coreKeys.courses.own()).toEqual(['own_courses'])
    expect(coreKeys.courses.templates()).toEqual(['template-courses'])
    expect(coreKeys.courses.copyability(COURSE)).toEqual(['course-copyability', COURSE])
    expect(coreKeys.courses.staff(COURSE)).toEqual(['courseStaff', COURSE])
    expect(coreKeys.courses.myParticipation(COURSE)).toEqual(['course_participation', COURSE])
  })

  it('keeps a missing id in the key rather than coercing it', () => {
    expect(coreKeys.coursePhases.byId(undefined)).toEqual(['course_phase', undefined])
    expect(coreKeys.applications.form(undefined)).toEqual(['application_form', undefined])
  })

  it('keeps the two course phase caches apart, since they hold the same phase', () => {
    expect(coreKeys.coursePhases.byId(PHASE)).toEqual(['course_phase', PHASE])
    expect(coreKeys.coursePhases.byIdInArchiveDialog(PHASE)).toEqual(['coursePhase', PHASE])
  })

  it('builds the course phase type keys', () => {
    expect(coreKeys.coursePhases.types()).toEqual(['course_phase_types'])
    expect(coreKeys.coursePhases.typesForScope('self')).toEqual(['coursePhaseType', 'self'])
    expect(coreKeys.coursePhases.typesForScope('all')).toEqual(['coursePhaseType', 'all'])
  })

  it('puts the three configurator graphs under one shared prefix', () => {
    expect(coreKeys.courseGraphs.phase(COURSE)).toEqual([
      'course_phases',
      'course_phase_graph',
      COURSE,
    ])
    expect(coreKeys.courseGraphs.participationData(COURSE)).toEqual([
      'course_phases',
      'participation_phase_graph',
      COURSE,
    ])
    expect(coreKeys.courseGraphs.phaseData(COURSE)).toEqual([
      'course_phases',
      'phase_phase_graph',
      COURSE,
    ])
  })

  it('builds the application keys as a prefix hierarchy', () => {
    expect(coreKeys.applications.all()).toEqual(['application'])
    expect(coreKeys.applications.inPhase(PHASE)).toEqual(['application', PHASE])
    expect(coreKeys.applications.ofParticipant(PHASE, PARTICIPATION)).toEqual([
      'application',
      PHASE,
      PARTICIPATION,
    ])
  })

  it('gives the staff dialog the same shape as a phase-scoped application', () => {
    expect(coreKeys.applications.ofParticipation(PARTICIPATION)).toEqual([
      'application',
      PARTICIPATION,
    ])
  })

  it('builds the participation keys as a prefix hierarchy', () => {
    expect(coreKeys.applications.participations.all()).toEqual(['application_participations'])
    expect(coreKeys.applications.participations.inPhase(PHASE)).toEqual([
      'application_participations',
      PHASE,
    ])
    expect(coreKeys.applications.participations.students(PHASE)).toEqual([
      'application_participations',
      'students',
      PHASE,
    ])
  })

  it('keeps the two application form caches apart, since they read different endpoints', () => {
    expect(coreKeys.applications.form(PHASE)).toEqual(['application_form', PHASE])
    expect(coreKeys.apply.form(PHASE)).toEqual(['applicationForm', PHASE])
  })

  it('builds the remaining application keys', () => {
    expect(coreKeys.applications.exportedAnswers(PHASE)).toEqual([
      'application_exported_answers',
      PHASE,
    ])
    expect(coreKeys.applications.universityUsers('ab', PHASE)).toEqual([
      'university_users',
      'ab',
      PHASE,
    ])
    expect(coreKeys.apply.open()).toEqual(['open_applications'])
  })

  it('builds the mail campaign keys', () => {
    expect(coreKeys.mailCampaigns.inCourse(COURSE)).toEqual(['mailCampaigns', COURSE])
    expect(coreKeys.mailCampaigns.byId(COURSE, CAMPAIGN)).toEqual([
      'mailCampaign',
      COURSE,
      CAMPAIGN,
    ])
    expect(coreKeys.mailCampaigns.recipientPreview(COURSE, CAMPAIGN)).toEqual([
      'mailCampaignRecipientPreview',
      COURSE,
      CAMPAIGN,
    ])
  })

  it('makes a scoped user search a descendant of the unscoped one', () => {
    expect(coreKeys.keycloak.userSearch.all()).toEqual(['keycloakUserSearch'])
    expect(coreKeys.keycloak.userSearch.forQuery('ab')).toEqual(['keycloakUserSearch', 'ab'])
    expect(coreKeys.keycloak.status()).toEqual(['keycloakStatus'])
  })

  it('builds the instructor note and student keys', () => {
    expect(coreKeys.instructorNotes.tags()).toEqual(['noteTags'])
    expect(coreKeys.instructorNotes.ofStudent(STUDENT)).toEqual(['instructorNotes', STUDENT])
    expect(coreKeys.students.byId(STUDENT)).toEqual(['student', STUDENT])
    expect(coreKeys.students.enrollments(STUDENT)).toEqual(['studentEnrollments', STUDENT])
  })

  it('builds the privacy keys, with the resource folded into the second element', () => {
    expect(coreKeys.privacy.latest('data-export')).toEqual(['privacy', 'data-export-latest'])
    expect(coreKeys.privacy.create('data-export')).toEqual(['privacy', 'data-export-create'])
    expect(coreKeys.privacy.status('data-deletion', 'request-1')).toEqual([
      'privacy',
      'data-deletion-status',
      'request-1',
    ])
    expect(coreKeys.privacy.admin.exports()).toEqual(['privacy', 'admin', 'exports'])
    expect(coreKeys.privacy.admin.deletions()).toEqual(['privacy', 'admin', 'deletions'])
  })

  it('names each service info cache in one element, so core is not their prefix', () => {
    expect(coreKeys.serviceInfo.core()).toEqual(['serviceInfo-core'])
    expect(coreKeys.serviceInfo.ofService('assessment')).toEqual(['serviceInfo-assessment'])
  })

  it('carries the audit log paging inputs in the key, so each page caches separately', () => {
    const filters = { outcome: 'denied' }
    const cursor = { createdAt: '2026-08-28T12:23:28Z', id: 'entry-1' }

    expect(coreKeys.auditLog.inCourse(COURSE, filters, 50, cursor)).toEqual([
      'auditLog',
      COURSE,
      filters,
      50,
      cursor,
    ])
    expect(coreKeys.auditLog.global(filters, 50, undefined)).toEqual([
      'auditLog',
      'global',
      filters,
      50,
      undefined,
    ])
    expect(coreKeys.auditLog.status()).toEqual(['auditLogStatus'])
  })

  it('builds the pull request banner key', () => {
    expect(coreKeys.githubPullRequest('2099')).toEqual(['github-pr', '2099'])
  })
})
