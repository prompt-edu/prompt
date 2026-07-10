import type { ProviderType } from '../../interfaces/providerConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const deleteProviderConfig = async (
  coursePhaseID: string,
  providerType: ProviderType,
): Promise<void> => {
  try {
    await infrastructureSetupAxiosInstance.delete(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/provider-configs/${providerType}`,
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
