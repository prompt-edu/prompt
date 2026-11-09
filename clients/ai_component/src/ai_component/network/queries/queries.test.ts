import type { InternalAxiosRequestConfig } from 'axios'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { aiAxiosInstance } from '../aiServerConfig'
import { getAICall } from './getAICall'
import { getAICalls } from './getAICalls'

const PHASE = 'phase-1'
const PHASE_PATH = `/ai/api/course_phase/${PHASE}`

let captured: InternalAxiosRequestConfig[]
const originalAdapter = aiAxiosInstance.defaults.adapter

beforeEach(() => {
  captured = []
  aiAxiosInstance.defaults.adapter = async (config) => {
    captured.push(config)
    return { data: {}, status: 200, statusText: 'OK', headers: {}, config }
  }
})

afterEach(() => {
  aiAxiosInstance.defaults.adapter = originalAdapter
})

describe('AI server queries', () => {
  it('lists the calls of a phase from the newest page', async () => {
    await getAICalls(PHASE, 50)
    expect(captured[0].url).toBe(`${PHASE_PATH}/calls`)
    expect(captured[0].params).toEqual({ limit: 50 })
  })

  it('passes the cursor of an older page', async () => {
    await getAICalls(PHASE, 50, { requestedAt: '2026-10-04T10:00:00Z', id: 'call-1' })
    expect(captured[0].params).toEqual({
      limit: 50,
      cursorRequestedAt: '2026-10-04T10:00:00Z',
      cursorId: 'call-1',
    })
  })

  it('reads one call through its own phase', async () => {
    await getAICall(PHASE, 'call-1')
    expect(captured[0].url).toBe(`${PHASE_PATH}/calls/call-1`)
  })
})
