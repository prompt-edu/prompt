import type {
  AICallCursor,
  AICallDetail,
  AICallPage,
  AIServiceInfo,
  AIStatus,
  PhaseKeyStatus,
} from '@core/interfaces/ai'
import { API_PREFIX, aiRequest, coreRequest } from '../client'

const phasePath = (phaseId: string) => `/ai/api/course_phase/${phaseId}`

// An AI server that is switched off or unreachable is an expected answer, not a failure.
const UNAVAILABLE = [404, 502, 503, 504]

export const ai = {
  status: (): Promise<AIStatus> => coreRequest.get(`${API_PREFIX}/ai/status`),

  info: (): Promise<AIServiceInfo> => aiRequest.get('/ai/api/info', { quietStatuses: UNAVAILABLE }),

  key: (phaseId: string): Promise<PhaseKeyStatus> => aiRequest.get(`${phasePath(phaseId)}/key`),

  setKey: (phaseId: string, key: string): Promise<PhaseKeyStatus> =>
    aiRequest.put(`${phasePath(phaseId)}/key`, { key }),

  removeKey: (phaseId: string): Promise<void> => aiRequest.del(`${phasePath(phaseId)}/key`),

  calls: (phaseId: string, limit: number, cursor?: AICallCursor): Promise<AICallPage> =>
    aiRequest.get(`${phasePath(phaseId)}/calls`, {
      params: cursor
        ? { limit, cursorRequestedAt: cursor.requestedAt, cursorId: cursor.id }
        : { limit },
    }),

  call: (phaseId: string, callId: string): Promise<AICallDetail> =>
    aiRequest.get(`${phasePath(phaseId)}/calls/${callId}`),
}
