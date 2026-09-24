import type { MyResource } from '../../interfaces/myResource'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getMyResources = async (coursePhaseID: string): Promise<MyResource[]> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.get(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/my-resources`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
