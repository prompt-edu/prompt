import type { AICall } from './aiCall'

export interface AICallCursor {
  requestedAt: string
  id: string
}

export interface AICallPage {
  calls: AICall[]
  nextCursor: AICallCursor | null
}
