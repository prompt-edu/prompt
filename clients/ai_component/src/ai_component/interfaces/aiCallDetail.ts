import type { ChatCompletionMessageParam } from 'openai/resources/chat/completions'
import type { AICall } from './aiCall'
import type { AICallEvent } from './aiCallEvent'

export enum AICallContentState {
  AVAILABLE = 'available',
  RESTRICTED = 'restricted',
  UNAVAILABLE = 'unavailable',
}

export interface AICallToolCall {
  id: string
  name: string
  arguments: string
}

export interface AICallChoice {
  index: number
  content: string
  refusal?: string
  toolCalls?: AICallToolCall[]
}

export interface AICallContent {
  request: { messages?: ChatCompletionMessageParam[] }
  choices: AICallChoice[] | null
  providerResponse?: string
}

export interface AICallDetail extends AICall {
  contentState: AICallContentState
  content?: AICallContent
  subjects: string[]
  events: AICallEvent[]
}
