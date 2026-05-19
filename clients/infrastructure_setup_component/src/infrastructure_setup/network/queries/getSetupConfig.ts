import { SetupConfig } from '../../interfaces/setupConfig'
import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const getSetupConfig = async (coursePhaseID: string): Promise<SetupConfig> => {
  try {
    return (
      await infrastructureSetupAxiosInstance.get(
        `/infrastructure-setup/api/course_phase/${coursePhaseID}/setup-config`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
