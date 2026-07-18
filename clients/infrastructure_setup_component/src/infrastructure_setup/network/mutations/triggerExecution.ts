import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const triggerExecution = async (coursePhaseID: string): Promise<void> => {
  try {
    await infrastructureSetupAxiosInstance.post(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/execute`,
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
