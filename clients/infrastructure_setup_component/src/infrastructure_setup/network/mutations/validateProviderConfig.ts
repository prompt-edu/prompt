import type { ProviderType } from '../../interfaces/providerConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const validateProviderConfig = async (
  coursePhaseID: string,
  providerType: ProviderType,
): Promise<void> => {
  await infrastructureSetupAxiosInstance.post(
    `/infrastructure-setup/api/course_phase/${coursePhaseID}/provider-configs/${providerType}/validate`,
  )
}
