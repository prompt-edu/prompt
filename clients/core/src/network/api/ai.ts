import type { AIServiceInfo } from '@core/managementConsole/ai/interfaces/aiServiceInfo'
import { aiRequest, NO_ANSWER } from '../client'

// An AI server that is switched off or unreachable is an expected answer, not a failure.
const UNAVAILABLE = [NO_ANSWER, 404, 502, 503, 504]

export const ai = {
  info: (): Promise<AIServiceInfo> => aiRequest.get('/ai/api/info', { quietStatuses: UNAVAILABLE }),
}
