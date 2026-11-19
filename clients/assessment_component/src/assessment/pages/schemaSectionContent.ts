import { AssessmentType } from '../interfaces/assessmentType'
import type { TutorLabel } from './hooks/useTutorLabel'

export interface SchemaSectionContent {
  title: string
  // The title as it reads mid-sentence; only the built-in words are lowercased.
  inlineTitle: string
  summary: string
  toggleLabel: string
  toggleHint: string
  schemaLabel: string
  schemaHint: string
  timeframeLabel: string
  timeframeHint: string
  lockHint: string
  detailTitle: string
  detailDescription: string
}

export const getSchemaSectionContent = (
  tutorLabel: TutorLabel,
): Record<AssessmentType, SchemaSectionContent> => ({
  [AssessmentType.ASSESSMENT]: {
    title: 'Assessment',
    inlineTitle: 'assessment',
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
    lockHint:
      "Assessments were already submitted with this schema. To keep them from being falsified, you can't turn the assessment off, switch to another schema or edit this one. To unlock it, open each participant's assessment page and delete the summary, the competency scores and the action items, and clear the category comments.",
    detailTitle: 'Assessment Schema',
    detailDescription:
      'Review and edit categories, competencies, weights, and proficiency descriptions for the final assessment.',
  },
  [AssessmentType.SELF]: {
    title: 'Self-Evaluation',
    inlineTitle: 'self-evaluation',
    summary:
      'Configure the reflective self-evaluation students complete before their final assessment is reviewed.',
    toggleLabel: 'Self-evaluation enabled',
    toggleHint: 'Turn this on when students should submit a structured reflection for this phase.',
    schemaLabel: 'Self-evaluation schema',
    schemaHint:
      'Choose the schema students use for self-reflection. Use the detail page to inspect or adapt its content.',
    timeframeLabel: 'Self-evaluation timeframe',
    timeframeHint: 'This controls when students can start and submit their self-evaluations.',
    lockHint:
      "Students already submitted self-evaluations with this schema. To keep them from being falsified, you can't turn the self-evaluation off, switch to another schema or edit this one.",
    detailTitle: 'Self-Evaluation Schema',
    detailDescription:
      'Review and edit the self-evaluation categories, competencies, weights, and proficiency descriptions.',
  },
  [AssessmentType.PEER]: {
    title: 'Peer-Evaluation',
    inlineTitle: 'peer-evaluation',
    summary:
      'Configure the peer feedback form used when students evaluate teammates or collaborators.',
    toggleLabel: 'Peer-evaluation enabled',
    toggleHint: 'Turn this on when peer feedback should be collected in this course phase.',
    schemaLabel: 'Peer-evaluation schema',
    schemaHint:
      'Choose the schema used for peer feedback. Shared schemas stay safe because edits create local copies when needed.',
    timeframeLabel: 'Peer-evaluation timeframe',
    timeframeHint: 'This controls when peer feedback opens and when submissions are due.',
    lockHint:
      "Students already submitted peer-evaluations with this schema. To keep them from being falsified, you can't turn the peer-evaluation off, switch to another schema or edit this one.",
    detailTitle: 'Peer-Evaluation Schema',
    detailDescription:
      'Review and edit the peer-evaluation categories, competencies, weights, and proficiency descriptions.',
  },
  [AssessmentType.TUTOR]: {
    title: `${tutorLabel.title}-Evaluation`,
    inlineTitle: `${tutorLabel.text}-evaluation`,
    summary: `Configure the feedback form students use to evaluate their ${tutorLabel.text} in this course phase.`,
    toggleLabel: `${tutorLabel.title}-evaluation enabled`,
    toggleHint: `Turn this on when students should submit structured feedback about their ${tutorLabel.text} in this phase.`,
    schemaLabel: `${tutorLabel.title}-evaluation schema`,
    schemaHint: `Choose the schema students use to evaluate their ${tutorLabel.text}. Use the detail page to review categories and competency descriptions.`,
    timeframeLabel: `${tutorLabel.title}-evaluation timeframe`,
    timeframeHint: `This controls when students can evaluate their ${tutorLabel.text} and when submissions close.`,
    lockHint: `Students already submitted ${tutorLabel.text}-evaluations with this schema. To keep them from being falsified, you can't turn the ${tutorLabel.text}-evaluation off, switch to another schema or edit this one.`,
    detailTitle: `${tutorLabel.title}-Evaluation Schema`,
    detailDescription: `Review and edit the ${tutorLabel.text}-evaluation categories, competencies, weights, and proficiency descriptions students use to evaluate their ${tutorLabel.text}.`,
  },
})
