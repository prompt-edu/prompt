import type { AICallDetail } from '../../interfaces/aiCallDetail'
import { aiAxiosInstance, aiPhasePath } from '../aiServerConfig'

// Every read is recorded as a content view of the call.
export const getAICall = async (coursePhaseId: string, callId: string): Promise<AICallDetail> =>
  (await aiAxiosInstance.get(`${aiPhasePath(coursePhaseId)}/calls/${callId}`)).data
