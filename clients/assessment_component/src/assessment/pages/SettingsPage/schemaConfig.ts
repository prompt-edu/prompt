import { AssessmentType } from '../../interfaces/assessmentType'
import { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'

export interface SchemaSectionContent {
  title: string
  summary: string
  toggleLabel: string
  toggleHint: string
  schemaLabel: string
  schemaHint: string
  timeframeLabel: string
  timeframeHint: string
  detailTitle: string
  detailDescription: string
}

export const schemaSectionContent: Record<AssessmentType, SchemaSectionContent> = {
  [AssessmentType.ASSESSMENT]: {
    title: 'Assessment',
    summary:
      'Configure the final assessment that lecturers use to grade students in this course phase.',
    toggleLabel: 'Assessment enabled',
    toggleHint: 'The main assessment is required for this phase and cannot be turned off.',
    schemaLabel: 'Assessment schema',
    schemaHint:
      'Choose the rubric used for the final assessment. Shared schemas are copied automatically before local edits.',
    timeframeLabel: 'Assessment timeframe',
    timeframeHint:
      'This controls when lecturers can submit final assessments for participants in this phase.',
    detailTitle: 'Assessment Schema',
    detailDescription:
      'Review and edit categories, competencies, weights, and proficiency descriptions for the final assessment.',
  },
  [AssessmentType.SELF]: {
    title: 'Self-Evaluation',
    summary:
      'Configure the reflective self-evaluation students complete before their final assessment is reviewed.',
    toggleLabel: 'Self-evaluation enabled',
    toggleHint: 'Turn this on when students should submit a structured reflection for this phase.',
    schemaLabel: 'Self-evaluation schema',
    schemaHint:
      'Choose the rubric students use for self-reflection. Use the detail page to inspect or adapt its content.',
    timeframeLabel: 'Self-evaluation timeframe',
    timeframeHint: 'This controls when students can start and submit their self-evaluations.',
    detailTitle: 'Self-Evaluation Schema',
    detailDescription:
      'Review and edit the self-evaluation categories, competencies, weights, and proficiency descriptions.',
  },
  [AssessmentType.PEER]: {
    title: 'Peer-Evaluation',
    summary:
      'Configure the peer feedback form used when students evaluate teammates or collaborators.',
    toggleLabel: 'Peer-evaluation enabled',
    toggleHint: 'Turn this on when peer feedback should be collected in this course phase.',
    schemaLabel: 'Peer-evaluation schema',
    schemaHint:
      'Choose the rubric used for peer feedback. Shared schemas stay safe because edits create local copies when needed.',
    timeframeLabel: 'Peer-evaluation timeframe',
    timeframeHint: 'This controls when peer feedback opens and when submissions are due.',
    detailTitle: 'Peer-Evaluation Schema',
    detailDescription:
      'Review and edit the peer-evaluation categories, competencies, weights, and proficiency descriptions.',
  },
  [AssessmentType.TUTOR]: {
    title: 'Tutor-Evaluation',
    summary:
      'Configure the rubric tutors use when they submit a dedicated evaluation in addition to the final assessment.',
    toggleLabel: 'Tutor-evaluation enabled',
    toggleHint:
      'Turn this on when tutors should complete a separate structured evaluation for this phase.',
    schemaLabel: 'Tutor-evaluation schema',
    schemaHint:
      'Choose the rubric tutors use. Use the detail page to review the categories and competency descriptions.',
    timeframeLabel: 'Tutor-evaluation timeframe',
    timeframeHint: 'This controls when tutor evaluations are available and when they close.',
    detailTitle: 'Tutor-Evaluation Schema',
    detailDescription:
      'Review and edit the tutor-evaluation categories, competencies, weights, and proficiency descriptions.',
  },
}

export const getSchemaIdForType = (
  config: CoursePhaseConfig | undefined,
  assessmentType: AssessmentType,
): string | undefined => {
  if (!config) return undefined

  switch (assessmentType) {
    case AssessmentType.SELF:
      return config.selfEvaluationSchema
    case AssessmentType.PEER:
      return config.peerEvaluationSchema
    case AssessmentType.TUTOR:
      return config.tutorEvaluationSchema
    case AssessmentType.ASSESSMENT:
    default:
      return config.assessmentSchemaID
  }
}

export const isSchemaTypeEnabled = (
  config: CoursePhaseConfig | undefined,
  assessmentType: AssessmentType,
): boolean => {
  if (!config) return false

  switch (assessmentType) {
    case AssessmentType.SELF:
      return config.selfEvaluationEnabled
    case AssessmentType.PEER:
      return config.peerEvaluationEnabled
    case AssessmentType.TUTOR:
      return config.tutorEvaluationEnabled
    case AssessmentType.ASSESSMENT:
    default:
      return true
  }
}
