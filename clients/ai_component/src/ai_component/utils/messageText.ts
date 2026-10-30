import type { ChatCompletionMessageParam } from 'openai/resources/chat/completions'

type ContentPart = Exclude<ChatCompletionMessageParam['content'], string | null | undefined>[number]

const isPart = (part: unknown): part is ContentPart =>
  typeof part === 'object' && part !== null && 'type' in part && typeof part.type === 'string'

const partText = (part: ContentPart): string => {
  switch (part.type) {
    case 'text':
      return part.text
    case 'refusal':
      return part.refusal
    default:
      return `[${part.type}]`
  }
}

// Stored requests are not validated, so a message without readable content is shown as raw JSON.
export const messageText = (message: ChatCompletionMessageParam): string => {
  const content = message?.content
  if (typeof content === 'string') return content
  if (Array.isArray(content) && content.every(isPart)) return content.map(partText).join('\n')
  return JSON.stringify(message, null, 2)
}
