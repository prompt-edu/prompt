import type { Team } from '@tumaet/prompt-shared-state'

import type { AssessmentParticipationWithStudent } from '../../../../interfaces/assessmentParticipationWithStudent'
import { AssessmentType } from '../../../../interfaces/assessmentType'
import type { CoursePhaseConfig } from '../../../../interfaces/coursePhaseConfig'
import type { EvaluationCompletion } from '../../../../interfaces/evaluationCompletion'
import { getReminderTypes } from './utils'

const baseConfig: CoursePhaseConfig = {
  coursePhaseID: 'phase-1',
  assessmentSchemaID: 'schema-assessment',
  start: new Date('2026-01-01T10:00:00Z'),
  deadline: new Date('2026-01-10T10:00:00Z'),
  selfEvaluationEnabled: false,
  selfEvaluationSchema: 'schema-self',
  selfEvaluationStart: new Date('2026-01-01T10:00:00Z'),
  selfEvaluationDeadline: new Date('2026-01-05T10:00:00Z'),
  peerEvaluationEnabled: false,
  peerEvaluationSchema: 'schema-peer',
  peerEvaluationStart: new Date('2026-01-01T10:00:00Z'),
  peerEvaluationDeadline: new Date('2026-01-06T10:00:00Z'),
  tutorEvaluationEnabled: false,
  tutorEvaluationSchema: 'schema-tutor',
  tutorEvaluationStart: new Date('2026-01-01T10:00:00Z'),
  tutorEvaluationDeadline: new Date('2026-01-07T10:00:00Z'),
  evaluationResultsVisible: false,
  gradeSuggestionVisible: false,
  actionItemsVisible: false,
  gradingSheetVisible: false,
  resultsReleased: false,
}

const assertDeepEqual = (actual: unknown, expected: unknown) => {
  const actualJson = JSON.stringify(actual)
  const expectedJson = JSON.stringify(expected)

  if (actualJson !== expectedJson) {
    throw new Error(`Expected ${expectedJson}, received ${actualJson}`)
  }
}

const buildParticipation = (
  courseParticipationID: string,
  teamID: string,
): AssessmentParticipationWithStudent =>
  ({
    coursePhaseID: 'phase-1',
    courseParticipationID,
    passStatus: 'NOT_ASSESSED',
    restrictedData: {},
    studentReadableData: {},
    prevData: {},
    student: {
      id: courseParticipationID,
      firstName: 'Test',
      lastName: 'Student',
      email: `${courseParticipationID}@example.com`,
      hasUniversityAccount: true,
    },
    teamID,
  }) as unknown as AssessmentParticipationWithStudent

const buildPerson = (id: string) => ({
  id,
  firstName: 'Test',
  lastName: 'Person',
})

const buildCompletion = (
  type: AssessmentType,
  authorCourseParticipationID: string,
  courseParticipationID: string,
  completed = true,
): EvaluationCompletion => ({
  type,
  authorCourseParticipationID,
  courseParticipationID,
  coursePhaseID: 'phase-1',
  completedAt: '2026-01-08T10:00:00Z',
  completed,
})

const testSingleMemberPeerTeam = () => {
  const reminderTypes = getReminderTypes(
    {
      ...baseConfig,
      peerEvaluationEnabled: true,
    },
    [buildParticipation('student-1', 'team-1')],
    [
      {
        id: 'team-1',
        name: 'Solo Team',
        members: [buildPerson('student-1')],
        tutors: [],
      } as unknown as Team,
    ],
    [],
  )

  assertDeepEqual(reminderTypes, [
    {
      type: 'peer',
      label: 'Peer Evaluation',
      deadline: baseConfig.peerEvaluationDeadline,
      recipientCount: 0,
    },
  ])
}

const testNoTutorRecipients = () => {
  const reminderTypes = getReminderTypes(
    {
      ...baseConfig,
      tutorEvaluationEnabled: true,
    },
    [buildParticipation('student-1', 'team-1'), buildParticipation('student-2', 'team-1')],
    [
      {
        id: 'team-1',
        name: 'No Tutor Team',
        members: [buildPerson('student-1'), buildPerson('student-2')],
        tutors: [],
      } as unknown as Team,
    ],
    [],
  )

  assertDeepEqual(reminderTypes, [
    {
      type: 'tutor',
      label: 'Tutor Evaluation',
      deadline: baseConfig.tutorEvaluationDeadline,
      recipientCount: 0,
    },
  ])
}

const testPartiallyCompletedPeerEvaluations = () => {
  const reminderTypes = getReminderTypes(
    {
      ...baseConfig,
      peerEvaluationEnabled: true,
    },
    [
      buildParticipation('student-1', 'team-1'),
      buildParticipation('student-2', 'team-1'),
      buildParticipation('student-3', 'team-1'),
    ],
    [
      {
        id: 'team-1',
        name: 'Three Person Team',
        members: [buildPerson('student-1'), buildPerson('student-2'), buildPerson('student-3')],
        tutors: [],
      } as unknown as Team,
    ],
    [
      buildCompletion(AssessmentType.PEER, 'student-1', 'student-2'),
      buildCompletion(AssessmentType.PEER, 'student-2', 'student-1'),
      buildCompletion(AssessmentType.PEER, 'student-2', 'student-3'),
      buildCompletion(AssessmentType.PEER, 'student-3', 'student-1'),
    ],
  )

  assertDeepEqual(reminderTypes, [
    {
      type: 'peer',
      label: 'Peer Evaluation',
      deadline: baseConfig.peerEvaluationDeadline,
      recipientCount: 2,
    },
  ])
}

testSingleMemberPeerTeam()
testNoTutorRecipients()
testPartiallyCompletedPeerEvaluations()
