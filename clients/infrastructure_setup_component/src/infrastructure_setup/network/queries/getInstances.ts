import type { ResourceInstance } from '../../interfaces/resourceInstance'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getInstances = async (coursePhaseID: string): Promise<ResourceInstance[]> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.get(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/instances`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
