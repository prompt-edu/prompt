import { infrastructureSetupAxiosInstance } from '../infrastructureSetupServerConfig'

// confirm=true is required by the server once the config has provisioned instances: the
// delete cascades to them, and those rows are PROMPT's only record of external resources
// it never deletes. The caller passes it after the lecturer confirmed in the dialog.
export const deleteResourceConfig = async (
  coursePhaseID: string,
  resourceConfigID: string,
  confirmed = false,
): Promise<void> => {
  try {
    await infrastructureSetupAxiosInstance.delete(
      `/infrastructure-setup/api/course_phase/${coursePhaseID}/resource-configs/${resourceConfigID}`,
      { params: { confirm: confirmed } },
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
