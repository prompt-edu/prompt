import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const deleteInstance = async (coursePhaseID: string, instanceID: string): Promise<void> => {
  await infrastructureSetupAxiosInstance.delete(
    `/infrastructure-setup/api/course_phase/${coursePhaseID}/instances/${instanceID}`,
  )
}
