import { ProviderConfig, UpsertProviderConfigRequest } from '../../interfaces/providerConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const upsertProviderConfig = async (
  coursePhaseID: string,
  request: UpsertProviderConfigRequest,
): Promise<ProviderConfig> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.put(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/provider-configs`,
        request,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
