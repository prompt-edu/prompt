export interface EvaluationOption {
  enabled: boolean
  schema: string
  start?: Date
  deadline?: Date
}

export interface EvaluationOptions {
  self: EvaluationOption
  peer: EvaluationOption
  tutor: EvaluationOption
}
