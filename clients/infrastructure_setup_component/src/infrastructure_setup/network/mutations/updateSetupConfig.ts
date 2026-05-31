import { SetupConfig, UpsertSetupConfigRequest } from '../../interfaces/setupConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const updateSetupConfig = async (
  coursePhaseID: string,
  request: UpsertSetupConfigRequest,
): Promise<SetupConfig> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.put(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/setup-config`,
        request,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
