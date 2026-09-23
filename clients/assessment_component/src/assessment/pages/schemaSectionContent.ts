import { AssessmentType } from '../interfaces/assessmentType'

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

export const getSchemaSectionContent = (
  tutorLabel: string,
): Record<AssessmentType, SchemaSectionContent> => ({
  [AssessmentType.ASSESSMENT]: {
    title: 'Assessment',
    summary:
      'Configure the final assessment that lecturers use to grade students in this course phase.',
    toggleLabel: 'Assessment enabled',
    toggleHint:
      'Turn this off to use the phase for evaluations only, without any lecturer grading. It locks once assessment data exists.',
    schemaLabel: 'Assessment schema',
    schemaHint:
      'Choose the schema used for the final assessment. Shared schemas are copied automatically before local edits.',
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
      'Choose the schema students use for self-reflection. Use the detail page to inspect or adapt its content.',
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
      'Choose the schema used for peer feedback. Shared schemas stay safe because edits create local copies when needed.',
    timeframeLabel: 'Peer-evaluation timeframe',
    timeframeHint: 'This controls when peer feedback opens and when submissions are due.',
    detailTitle: 'Peer-Evaluation Schema',
    detailDescription:
      'Review and edit the peer-evaluation categories, competencies, weights, and proficiency descriptions.',
  },
  [AssessmentType.TUTOR]: {
    title: `${tutorLabel}-Evaluation`,
    summary: `Configure the feedback form students use to evaluate their ${tutorLabel} in this course phase.`,
    toggleLabel: `${tutorLabel}-evaluation enabled`,
    toggleHint: `Turn this on when students should submit structured feedback about their ${tutorLabel} in this phase.`,
    schemaLabel: `${tutorLabel}-evaluation schema`,
    schemaHint: `Choose the schema students use to evaluate their ${tutorLabel}. Use the detail page to review categories and competency descriptions.`,
    timeframeLabel: `${tutorLabel}-evaluation timeframe`,
    timeframeHint: `This controls when students can evaluate their ${tutorLabel} and when submissions close.`,
    detailTitle: `${tutorLabel}-Evaluation Schema`,
    detailDescription: `Review and edit the ${tutorLabel}-evaluation categories, competencies, weights, and proficiency descriptions students use to evaluate their ${tutorLabel}.`,
  },
})
