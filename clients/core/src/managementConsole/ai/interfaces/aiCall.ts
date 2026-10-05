export enum AICallOutcome {
  PENDING = 'pending',
  SUCCESS = 'success',
  ERROR = 'error',
  TIMEOUT = 'timeout',
  CANCELLED = 'cancelled',
  DENIED = 'denied',
}

export interface AICall {
  id: string
  actorId: string
  actorRole: string
  feature: string
  template: string | null
  templateVersion: string | null
  requestedModel: string | null
  servedModel: string | null
  outcome: AICallOutcome
  httpStatus: number | null
  finishReason: string | null
  errorCode: string | null
  promptTokens: number | null
  completionTokens: number | null
  streamed: boolean
  requestedAt: string
  firstTokenAt: string | null
  completedAt: string | null
}
