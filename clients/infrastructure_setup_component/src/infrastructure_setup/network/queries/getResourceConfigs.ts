import type { ResourceConfig } from '../../interfaces/resourceConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getResourceConfigs = async (coursePhaseID: string): Promise<ResourceConfig[]> => {
  return (
    await infrastructureSetupAxiosInstance.get(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/resource-configs`,
    )
  ).data
}
