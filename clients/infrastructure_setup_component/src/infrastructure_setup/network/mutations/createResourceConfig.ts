import type { CreateResourceConfigRequest, ResourceConfig } from '../../interfaces/resourceConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const createResourceConfig = async (
  coursePhaseID: string,
  request: CreateResourceConfigRequest,
): Promise<ResourceConfig> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.post(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/resource-configs`,
        request,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
