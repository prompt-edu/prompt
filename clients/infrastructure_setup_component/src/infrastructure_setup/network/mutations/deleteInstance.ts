import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const deleteInstance = async (coursePhaseID: string, instanceID: string): Promise<void> => {
  try {
    await infrastructureSetupAxiosInstance.delete(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/instances/${instanceID}`,
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
