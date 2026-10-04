import { randomUUID } from 'node:crypto'
import { apiContextFor, expect, test } from '../../src/fixtures/api'
import { ASSESSMENT_FOREIGN_PHASE_ID, FULL_COURSE_PHASES } from '../../src/data/constants'
import { aiUrl, complete, providerRequests, setPhaseKey, SUMMARY_ANSWER } from './helpers'

const PHASE = FULL_COURSE_PHASES.assessment.id

test.describe('AI server gateway', () => {
  test.beforeAll(async () => {
    const lecturer = await apiContextFor('course-lecturer')
    try {
      await setPhaseKey(lecturer, PHASE, 'logos-e2e-key-gateway')
    } finally {
      await lecturer.dispose()
    }
  })

  test('streams a completion through the AI server and records it', async ({ apiAs }) => {
    const lecturer = await apiAs('course-lecturer')
    const marker = randomUUID()

    const res = await complete(lecturer, PHASE, {
      stream: true,
      marker,
      headers: {
        'X-Prompt-Feature': 'assessment.action_item_suggestions',
        'X-Prompt-Template': 'e2e',
        'X-Prompt-Template-Version': '1',
      },
    })

    expect(res.status()).toBe(200)
    expect(res.headers()['content-type']).toContain('text/event-stream')
    const stream = await res.text()
    expect(stream).toContain('data: [DONE]')
    const callId = res.headers()['x-prompt-ai-call-id']
    expect(callId).toBeTruthy()

    const [forwarded] = await providerRequests(lecturer, marker)
    expect(forwarded.body).not.toHaveProperty('user')
    expect(forwarded.body.stream_options).toEqual({ include_usage: true })

    const admin = await apiAs('admin')
    const call = await admin.get(aiUrl(PHASE, `calls/${callId}`))
    expect(call.status()).toBe(200)
    const detail = (await call.json()) as {
      outcome: string
      feature: string
      promptTokens: number
      content: { responseText: string }
    }
    expect(detail.outcome).toBe('success')
    expect(detail.feature).toBe('assessment.action_item_suggestions')
    expect(detail.promptTokens).toBe(42)
    expect(detail.content.responseText).toBe(SUMMARY_ANSWER)
  })

  test('keeps audit reads to admins', async ({ apiAs }) => {
    const lecturer = await apiAs('course-lecturer')
    expect((await lecturer.get(aiUrl(PHASE, 'calls'))).status()).toBe(403)

    const admin = await apiAs('admin')
    expect((await admin.get(aiUrl(PHASE, 'calls'))).status()).toBe(200)
  })

  test('rejects students and a PromptLecturer without a role in the course', async ({ apiAs }) => {
    const student = await apiAs('student')
    expect((await complete(student, PHASE)).status()).toBe(403)
    expect((await student.get(aiUrl(PHASE, 'key'))).status()).toBe(403)

    // `lecturer` holds PROMPT_Lecturer, which the AI server never accepts on its own.
    const lecturer = await apiAs('lecturer')
    expect((await complete(lecturer, ASSESSMENT_FOREIGN_PHASE_ID)).status()).toBe(403)
  })

  test('never returns the key', async ({ apiAs }) => {
    const lecturer = await apiAs('course-lecturer')
    const res = await lecturer.get(aiUrl(PHASE, 'key'))
    expect(res.status()).toBe(200)
    const body = await res.text()
    expect(body).not.toContain('logos-e2e-key-gateway')
    expect(JSON.parse(body)).toMatchObject({ configured: true, last4: 'eway' })
  })
})
