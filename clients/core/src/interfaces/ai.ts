export interface AIStatus {
  enabled: boolean
}

export interface AIServiceInfo {
  serviceName?: string
  healthy?: boolean
}

export interface PhaseKeyStatus {
  configured: boolean
  last4?: string
  setBy?: string
  setAt?: string
}

export interface AICallCursor {
  requestedAt: string
  id: string
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
  outcome: 'pending' | 'success' | 'error' | 'timeout' | 'cancelled' | 'denied'
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

export interface AICallPage {
  calls: AICall[]
  nextCursor: AICallCursor | null
}

export interface AICallEvent {
  id: string
  actorId: string
  type: 'shown' | 'accepted' | 'edited' | 'rejected' | 'content_viewed'
  data: Record<string, unknown>
  createdAt: string
}

export interface AICallMessage {
  role: string
  content: string
}

export interface AICallDetail extends AICall {
  contentState: 'available' | 'restricted' | 'unavailable'
  content?: {
    request: { messages?: AICallMessage[] }
    responseText: string
  }
  subjects: string[]
  events: AICallEvent[]
}
