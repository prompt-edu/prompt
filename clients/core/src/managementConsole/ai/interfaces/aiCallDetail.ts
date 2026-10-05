import type { AICall } from './aiCall'
import type { AICallEvent } from './aiCallEvent'

export enum AICallContentState {
  AVAILABLE = 'available',
  RESTRICTED = 'restricted',
  UNAVAILABLE = 'unavailable',
}

export interface AICallMessage {
  role: string
  content: string
}

export interface AICallContent {
  request: { messages?: AICallMessage[] }
  responseText: string
}

export interface AICallDetail extends AICall {
  contentState: AICallContentState
  content?: AICallContent
  subjects: string[]
  events: AICallEvent[]
}
