import { randomUUID } from 'node:crypto'
import { expect, test } from '../../src/fixtures/api'
import { ASSESSMENT_FOREIGN_PHASE_ID, FULL_COURSE_PHASES } from '../../src/data/constants'
import { aiUrl, complete, FEATURE, PROVIDER_KEY, providerRequests, SUMMARY_ANSWER } from './helpers'

const PHASE = FULL_COURSE_PHASES.assessment.id

test.describe('AI server gateway', () => {
  test('streams a completion through the AI server and records it', async ({ apiAs }) => {
    const lecturer = await apiAs('course-lecturer')
    const marker = randomUUID()

    const res = await complete(lecturer, PHASE, {
      stream: true,
      marker,
      headers: { 'X-Prompt-Template': 'e2e', 'X-Prompt-Template-Version': '1' },
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
    const body = await call.text()
    expect(body).not.toContain(PROVIDER_KEY)
    const detail = JSON.parse(body) as {
      outcome: string
      feature: string
      issuer: string
      promptTokens: number
      content: { responseText: string }
    }
    expect(detail.outcome).toBe('success')
    expect(detail.feature).toBe(FEATURE)
    expect(detail.issuer).toMatch(/\/realms\//)
    expect(detail.promptTokens).toBe(42)
    expect(detail.content.responseText).toBe(SUMMARY_ANSWER)
  })

  test('refuses a call without the phase key or a feature', async ({ apiAs }) => {
    const lecturer = await apiAs('course-lecturer')

    const withoutKey = await complete(lecturer, PHASE, { headers: { 'X-Prompt-Provider-Key': '' } })
    expect(withoutKey.status()).toBe(400)
    const withoutFeature = await complete(lecturer, PHASE, { headers: { 'X-Prompt-Feature': '' } })
    expect(withoutFeature.status()).toBe(400)
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

    // `lecturer` holds PROMPT_Lecturer, which the AI server never accepts on its own.
    const lecturer = await apiAs('lecturer')
    expect((await complete(lecturer, ASSESSMENT_FOREIGN_PHASE_ID)).status()).toBe(403)
  })
})
