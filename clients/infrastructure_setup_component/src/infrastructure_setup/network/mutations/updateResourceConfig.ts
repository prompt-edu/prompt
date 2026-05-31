import { ResourceConfig, UpdateResourceConfigRequest } from '../../interfaces/resourceConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const updateResourceConfig = async (
  coursePhaseID: string,
  resourceConfigID: string,
  request: UpdateResourceConfigRequest,
): Promise<ResourceConfig> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.put(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/resource-configs/${resourceConfigID}`,
        request,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
