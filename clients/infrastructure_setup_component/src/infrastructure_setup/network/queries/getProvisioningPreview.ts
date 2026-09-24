import type { ProvisioningPreview } from '../../interfaces/provisioningPreview'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getProvisioningPreview = async (
  coursePhaseID: string,
): Promise<ProvisioningPreview> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.get(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/execute/preview`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
