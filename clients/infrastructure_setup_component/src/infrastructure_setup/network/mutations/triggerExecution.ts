import type { TriggerSummary } from '../../interfaces/triggerSummary'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const triggerExecution = async (coursePhaseID: string): Promise<TriggerSummary> => {
  return (
    await infrastructureSetupAxiosInstance.post<TriggerSummary>(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/execute`,
    )
  ).data
}
