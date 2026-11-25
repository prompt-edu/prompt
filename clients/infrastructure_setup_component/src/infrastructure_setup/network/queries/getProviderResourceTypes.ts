import type { ProviderType } from '../../interfaces/providerConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getProviderResourceTypes = async (
  coursePhaseID: string,
  providerType: ProviderType,
): Promise<string[]> => {
  return (
    await infrastructureSetupAxiosInstance.get(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/provider-configs/${providerType}/resource-types`,
    )
  ).data
}
