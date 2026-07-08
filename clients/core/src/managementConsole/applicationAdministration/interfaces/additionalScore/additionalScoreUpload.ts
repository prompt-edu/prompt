import type { IndividualScore } from './individualScore'

export interface AdditionalScoreUpload {
  name: string
  key: string
  threshold: number
  thresholdActive: boolean
  scores: IndividualScore[]
}
