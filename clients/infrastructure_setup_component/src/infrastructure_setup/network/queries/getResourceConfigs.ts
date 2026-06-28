import { ResourceConfig } from '../../interfaces/resourceConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getResourceConfigs = async (coursePhaseID: string): Promise<ResourceConfig[]> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.get(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/resource-configs`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
