import type { AICallCursor, AICallPage } from '../../interfaces/aiCallPage'
import { aiAxiosInstance, aiPhasePath } from '../aiServerConfig'

export const getAICalls = async (
  coursePhaseId: string,
  limit: number,
  cursor?: AICallCursor,
): Promise<AICallPage> => {
  const params = cursor
    ? { limit, cursorRequestedAt: cursor.requestedAt, cursorId: cursor.id }
    : { limit }
  return (await aiAxiosInstance.get(`${aiPhasePath(coursePhaseId)}/calls`, { params })).data
}
