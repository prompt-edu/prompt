import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const retryInstance = async (coursePhaseID: string, instanceID: string): Promise<void> => {
  try {
    await infrastructureSetupAxiosInstance.post(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/instances/${instanceID}/retry`,
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
