export enum AssessmentType {
  SELF = 'self',
  PEER = 'peer',
  TUTOR = 'tutor',
  ASSESSMENT = 'assessment',
}

export type EvaluationType = AssessmentType.SELF | AssessmentType.PEER | AssessmentType.TUTOR
