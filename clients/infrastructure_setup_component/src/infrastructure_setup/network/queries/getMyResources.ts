import type { MyResource } from '../../interfaces/myResource'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getMyResources = async (coursePhaseID: string): Promise<MyResource[]> => {
  return (
    await infrastructureSetupAxiosInstance.get(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/my-resources`,
    )
  ).data
}
