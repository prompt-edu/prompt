import type { TriggerSummary } from '../../interfaces/triggerSummary'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const triggerExecution = async (coursePhaseID: string): Promise<TriggerSummary> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.post<TriggerSummary>(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/execute`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
