import { axiosInstance } from '@tumaet/prompt-shared-state'
import { UpdateCoursePhaseParticipationStatus } from '@tumaet/prompt-shared-state'

export const updateApplicationStatus = async (
  coursePhaseID: string,
  updateApplications: UpdateCoursePhaseParticipationStatus,
): Promise<void> => {
  try {
    await axiosInstance.put(`/api/applications/${coursePhaseID}/assessment`, updateApplications, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    })
  } catch (err) {
    console.error(err)
    throw err
  }
}
