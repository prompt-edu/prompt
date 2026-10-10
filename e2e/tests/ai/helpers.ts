import { APIRequestContext, expect } from '@playwright/test'
import { AI_API, AIMOCK_URL, BASE_URL } from '../../src/env'

// Matches a fixture in e2e/fixtures/aimock, which the AI server's Go tests share.
export const TEST_MODEL = 'prompt-test-model'
export const SUMMARY_PROMPT = 'Summarize the peer feedback'
export const SUMMARY_ANSWER = 'The team communicates well and should record its decisions earlier.'

export const FEATURE = 'assessment.action_item_suggestions'
// Stands in for the key a phase server keeps for its phase; the AI server keeps none.
export const PROVIDER_KEY = 'logos-e2e-phase-key'

export const aiUrl = (phaseId: string, path: string) =>
  `${BASE_URL}${AI_API}/course_phase/${phaseId}/${path}`

// The headers a phase server sends. An empty value leaves that header out.
export const phaseHeaders = (overrides: Record<string, string> = {}): Record<string, string> =>
  Object.fromEntries(
    Object.entries({
      'X-Prompt-Provider-Key': PROVIDER_KEY,
      'X-Prompt-Feature': FEATURE,
      ...overrides,
    }).filter(([, value]) => value !== ''),
  )

export async function complete(
  api: APIRequestContext,
  phaseId: string,
  options: { stream?: boolean; headers?: Record<string, string>; marker?: string } = {},
) {
  return api.post(aiUrl(phaseId, 'v1/chat/completions'), {
    headers: phaseHeaders(options.headers),
    data: {
      model: TEST_MODEL,
      stream: options.stream ?? false,
      user: 'student@example.com',
      messages: [{ role: 'user', content: `${SUMMARY_PROMPT} ${options.marker ?? ''}` }],
    },
  })
}

interface JournalEntry {
  path: string
  body: Record<string, unknown>
}

// The requests aimock received that carry the marker; the shard's specs share one aimock.
export async function providerRequests(
  api: APIRequestContext,
  marker: string,
): Promise<JournalEntry[]> {
  const res = await api.get(`${AIMOCK_URL}/__aimock/journal`)
  expect(res.ok()).toBe(true)
  const entries = (await res.json()) as JournalEntry[]
  return entries.filter((entry) => JSON.stringify(entry.body).includes(marker))
}
