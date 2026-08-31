import { describe, expect, it } from 'vitest'

import { AssessmentType } from '../../interfaces/assessmentType'
import { assessmentKeys } from './keys'

const PHASE = 'phase-1'
const PARTICIPATION = 'participation-1'
const SCHEMA = 'schema-1'

describe('assessmentKeys', () => {
  it('builds the assessment keys as a prefix hierarchy', () => {
    expect(assessmentKeys.assessments.all()).toEqual(['assessments'])
    expect(assessmentKeys.assessments.inPhase(PHASE)).toEqual(['assessments', PHASE])
    expect(assessmentKeys.assessments.ofParticipant(PHASE, PARTICIPATION)).toEqual([
      'assessments',
      PHASE,
      PARTICIPATION,
    ])
  })

  it('keeps a missing id in the key rather than coercing it', () => {
    expect(assessmentKeys.assessments.inPhase(undefined)).toEqual(['assessments', undefined])
    expect(assessmentKeys.categories(undefined)).toEqual(['categories', undefined])
  })

  it('builds the schema keys', () => {
    expect(assessmentKeys.assessmentSchemas.inPhase(PHASE)).toEqual(['assessmentSchemas', PHASE])
    expect(assessmentKeys.assessmentSchemas.hasAssessmentData(SCHEMA, PHASE)).toEqual([
      'schemaHasAssessmentData',
      SCHEMA,
      PHASE,
    ])
  })

  it('names one category cache per evaluation type', () => {
    expect(assessmentKeys.categories(PHASE)).toEqual(['categories', PHASE])
    expect(assessmentKeys.evaluationCategories(AssessmentType.SELF, PHASE)).toEqual([
      'selfEvaluationCategories',
      PHASE,
    ])
    expect(assessmentKeys.evaluationCategories(AssessmentType.PEER, PHASE)).toEqual([
      'peerEvaluationCategories',
      PHASE,
    ])
    expect(assessmentKeys.evaluationCategories(AssessmentType.TUTOR, PHASE)).toEqual([
      'tutorEvaluationCategories',
      PHASE,
    ])
  })

  it('builds the evaluation keys, including the one whose type leads', () => {
    expect(assessmentKeys.evaluations.inPhase(PHASE)).toEqual(['evaluations', PHASE])
    expect(assessmentKeys.evaluations.mine(PHASE)).toEqual(['my-evaluations', PHASE])
    expect(assessmentKeys.evaluations.ofTutor(PHASE, PARTICIPATION)).toEqual([
      'tutor-evaluations',
      PHASE,
      PARTICIPATION,
    ])
    expect(
      assessmentKeys.evaluations.ofParticipant(AssessmentType.SELF, PHASE, PARTICIPATION),
    ).toEqual(['self', 'evaluations', PHASE, PARTICIPATION])
  })

  it('builds the completion keys', () => {
    expect(assessmentKeys.assessmentCompletions(PHASE)).toEqual(['assessmentCompletions', PHASE])
    expect(assessmentKeys.evaluationCompletions.inPhase(PHASE)).toEqual([
      'evaluationCompletions',
      PHASE,
    ])
    expect(assessmentKeys.evaluationCompletions.mine(PHASE)).toEqual([
      'my-evaluation-completions',
      PHASE,
    ])
  })

  it('builds the feedback item keys', () => {
    expect(assessmentKeys.feedbackItems.inPhase(PHASE)).toEqual(['all-feedback-items', PHASE])
    expect(assessmentKeys.feedbackItems.ofStudent(PHASE, PARTICIPATION)).toEqual([
      'student-feedback-items',
      PHASE,
      PARTICIPATION,
    ])
    expect(assessmentKeys.feedbackItems.ofTutor(PHASE, PARTICIPATION)).toEqual([
      'tutor-feedback-items',
      PHASE,
      PARTICIPATION,
    ])
    expect(assessmentKeys.feedbackItems.mine(PHASE)).toEqual(['my-feedback-items', PHASE])
  })

  it('builds the action item keys', () => {
    expect(assessmentKeys.actionItems.inPhase(PHASE)).toEqual(['actionItems', PHASE])
    expect(assessmentKeys.actionItems.ofParticipant(PHASE, PARTICIPATION)).toEqual([
      'actionItems',
      PHASE,
      PARTICIPATION,
    ])
    expect(assessmentKeys.actionItems.mine(PHASE)).toEqual(['myActionItems', PHASE])
  })

  it('builds the student result keys', () => {
    expect(assessmentKeys.results.myAssessment(PHASE)).toEqual(['myAssessmentResults', PHASE])
    expect(assessmentKeys.results.myEvaluation(PHASE)).toEqual(['myEvaluationResults', PHASE])
    expect(assessmentKeys.results.myGradeSuggestion(PHASE)).toEqual(['myGradeSuggestion', PHASE])
  })

  it('builds the remaining phase keys', () => {
    expect(assessmentKeys.scoreLevels(PHASE)).toEqual(['scoreLevels', PHASE])
    expect(assessmentKeys.coursePhaseConfig(PHASE)).toEqual(['coursePhaseConfig', PHASE])
    expect(assessmentKeys.teams(PHASE)).toEqual(['teams', PHASE])
    expect(assessmentKeys.myParticipation(PHASE)).toEqual(['course_phase_participation', PHASE])
  })

  it('reproduces the keys owned by the shared state package', () => {
    expect(assessmentKeys.participants(PHASE)).toEqual(['participants', PHASE])
    expect(assessmentKeys.coursePhase(PHASE)).toEqual(['course_phase', PHASE])
  })
})
