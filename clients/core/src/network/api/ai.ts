import type { AICallDetail } from '@core/managementConsole/ai/interfaces/aiCallDetail'
import type { AICallCursor, AICallPage } from '@core/managementConsole/ai/interfaces/aiCallPage'
import type { AIServiceInfo } from '@core/managementConsole/ai/interfaces/aiServiceInfo'
import type { PhaseKeyStatus } from '@core/managementConsole/ai/interfaces/phaseKeyStatus'
import { aiRequest, NO_ANSWER } from '../client'

const phasePath = (phaseId: string) => `/ai/api/course_phase/${phaseId}`

// An AI server that is switched off or unreachable is an expected answer, not a failure.
const UNAVAILABLE = [NO_ANSWER, 404, 502, 503, 504]

export const ai = {
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
