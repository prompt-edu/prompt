import { axiosInstance } from '@tumaet/prompt-shared-state'

export const deleteApplications = async (
  coursePhaseID: string,
  courseParticipationIDs: string[],
): Promise<void> => {
  try {
    return await axiosInstance.delete(`/api/applications/${coursePhaseID}`, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
      data: courseParticipationIDs,
    })
  } catch (err) {
    console.error(err)
    throw err
  }
}
