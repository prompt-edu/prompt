import type { ResourceConfig } from '../../interfaces/resourceConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getResourceConfig = async (
  coursePhaseID: string,
  resourceConfigID: string,
): Promise<ResourceConfig> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.get(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/resource-configs/${resourceConfigID}`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
