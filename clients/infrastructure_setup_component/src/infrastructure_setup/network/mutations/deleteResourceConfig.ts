import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

export const deleteResourceConfig = async (
  coursePhaseID: string,
  resourceConfigID: string,
): Promise<void> => {
  try {
    await infrastructureSetupAxiosInstance.delete(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/resource-configs/${resourceConfigID}`,
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
