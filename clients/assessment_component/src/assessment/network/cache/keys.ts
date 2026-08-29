import { AssessmentType, type EvaluationType } from '../../interfaces/assessmentType'

type Id = string | undefined

const EVALUATION_CATEGORY_KEY: Record<EvaluationType, string> = {
  [AssessmentType.SELF]: 'selfEvaluationCategories',
  [AssessmentType.PEER]: 'peerEvaluationCategories',
  [AssessmentType.TUTOR]: 'tutorEvaluationCategories',
}

export const assessmentKeys = {
  assessments: {
    all: () => ['assessments'] as const,
    inPhase: (phaseId: Id) => ['assessments', phaseId] as const,
    ofParticipant: (phaseId: Id, courseParticipationId: Id) =>
      ['assessments', phaseId, courseParticipationId] as const,
  },
  scoreLevels: (phaseId: Id) => ['scoreLevels', phaseId] as const,
  assessmentCompletions: (phaseId: Id) => ['assessmentCompletions', phaseId] as const,
  assessmentSchemas: {
    inPhase: (phaseId: Id) => ['assessmentSchemas', phaseId] as const,
    hasAssessmentData: (schemaId: Id, phaseId: Id) =>
      ['schemaHasAssessmentData', schemaId, phaseId] as const,
  },
  categories: (phaseId: Id) => ['categories', phaseId] as const,
  evaluationCategories: (evaluationType: EvaluationType, phaseId: Id) =>
    [EVALUATION_CATEGORY_KEY[evaluationType], phaseId] as const,
  coursePhaseConfig: (phaseId: Id) => ['coursePhaseConfig', phaseId] as const,
  actionItems: {
    inPhase: (phaseId: Id) => ['actionItems', phaseId] as const,
    ofParticipant: (phaseId: Id, courseParticipationId: Id) =>
      ['actionItems', phaseId, courseParticipationId] as const,
    mine: (phaseId: Id) => ['myActionItems', phaseId] as const,
  },
  teams: (phaseId: Id) => ['teams', phaseId] as const,
  evaluations: {
    inPhase: (phaseId: Id) => ['evaluations', phaseId] as const,
    mine: (phaseId: Id) => ['my-evaluations', phaseId] as const,
    // The type leads, so this key is not a descendant of `evaluations.inPhase`
    ofParticipant: (assessmentType: AssessmentType, phaseId: Id, courseParticipationId: Id) =>
      [assessmentType, 'evaluations', phaseId, courseParticipationId] as const,
    ofTutor: (phaseId: Id, tutorParticipationId: Id) =>
      ['tutor-evaluations', phaseId, tutorParticipationId] as const,
  },
  evaluationCompletions: {
    inPhase: (phaseId: Id) => ['evaluationCompletions', phaseId] as const,
    mine: (phaseId: Id) => ['my-evaluation-completions', phaseId] as const,
  },
  feedbackItems: {
    inPhase: (phaseId: Id) => ['all-feedback-items', phaseId] as const,
    ofStudent: (phaseId: Id, courseParticipationId: Id) =>
      ['student-feedback-items', phaseId, courseParticipationId] as const,
    ofTutor: (phaseId: Id, tutorParticipationId: Id) =>
      ['tutor-feedback-items', phaseId, tutorParticipationId] as const,
    mine: (phaseId: Id) => ['my-feedback-items', phaseId] as const,
  },
  results: {
    myAssessment: (phaseId: Id) => ['myAssessmentResults', phaseId] as const,
    myEvaluation: (phaseId: Id) => ['myEvaluationResults', phaseId] as const,
    myGradeSuggestion: (phaseId: Id) => ['myGradeSuggestion', phaseId] as const,
  },
  myParticipation: (phaseId: Id) => ['course_phase_participation', phaseId] as const,
  // Owned by @tumaet/prompt-shared-state: react-query is a Module Federation singleton, so these
  // two entries are shared with core and must keep their literals
  participants: (phaseId: Id) => ['participants', phaseId] as const,
  coursePhase: (phaseId: Id) => ['course_phase', phaseId] as const,
}
