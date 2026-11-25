import type { ResourceInstance } from '../../interfaces/resourceInstance'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getInstances = async (coursePhaseID: string): Promise<ResourceInstance[]> => {
  return (
    await infrastructureSetupAxiosInstance.get(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/instances`,
    )
  ).data
}
