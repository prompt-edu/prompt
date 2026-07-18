import type { AuthField, ProviderType } from '../../interfaces/providerConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getProviderAuthFields = async (
  coursePhaseID: string,
  providerType: ProviderType,
): Promise<AuthField[]> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.get(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/provider-configs/${providerType}/fields`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
