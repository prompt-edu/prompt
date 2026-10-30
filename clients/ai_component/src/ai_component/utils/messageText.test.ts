import type { ChatCompletionMessageParam } from 'openai/resources/chat/completions'
import { describe, expect, it } from 'vitest'
import { messageText } from './messageText'

describe('messageText', () => {
  it('returns string content as is', () => {
    expect(messageText({ role: 'user', content: 'Hello' })).toBe('Hello')
  })

  it('joins text and refusal parts and marks other parts by type', () => {
    expect(
      messageText({
        role: 'user',
        content: [
          { type: 'text', text: 'First' },
          { type: 'image_url', image_url: { url: 'https://example.com/a.png' } },
          { type: 'text', text: 'Second' },
        ],
      }),
    ).toBe('First\n[image_url]\nSecond')
    expect(messageText({ role: 'assistant', content: [{ type: 'refusal', refusal: 'No' }] })).toBe(
      'No',
    )
  })

  it('shows an assistant turn that only carries tool calls as raw JSON', () => {
    const toolCallTurn: ChatCompletionMessageParam = {
      role: 'assistant',
      tool_calls: [
        { id: 'call_1', type: 'function', function: { name: 'lookup', arguments: '{"id":1}' } },
      ],
    }
    expect(messageText(toolCallTurn)).toBe(JSON.stringify(toolCallTurn, null, 2))
  })

  it('shows content that does not match the format as raw JSON instead of throwing', () => {
    for (const stored of [
      { role: 'user', content: { text: 'not an array' } },
      { role: 'user', content: [null] },
      null,
    ]) {
      const message = stored as unknown as ChatCompletionMessageParam
      expect(messageText(message)).toBe(JSON.stringify(message, null, 2))
    }
  })
})
