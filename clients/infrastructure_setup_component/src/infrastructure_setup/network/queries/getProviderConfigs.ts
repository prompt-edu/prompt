import type { ProviderConfig } from '../../interfaces/providerConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getProviderConfigs = async (coursePhaseID: string): Promise<ProviderConfig[]> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.get(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/provider-configs`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
