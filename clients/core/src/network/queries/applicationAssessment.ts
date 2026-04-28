import { axiosInstance } from '@tumaet/prompt-shared-state'
import { GetApplication } from '../../interfaces/application/getApplication'

export const getApplicationAssessment = async (
  coursePhaseId: string,
  courseParticipationID,
): Promise<GetApplication> => {
  try {
    return (await axiosInstance.get(`/api/applications/${coursePhaseId}/${courseParticipationID}`))
      .data
  } catch (err) {
    console.error(err)
    throw err
  }
}
