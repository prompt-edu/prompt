import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const retryInstance = async (coursePhaseID: string, instanceID: string): Promise<void> => {
  await infrastructureSetupAxiosInstance.post(
    `/infrastructure-setup/api/course_phase/${coursePhaseID}/instances/${instanceID}/retry`,
  )
}
