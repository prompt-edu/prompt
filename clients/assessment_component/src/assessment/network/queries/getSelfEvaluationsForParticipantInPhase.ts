import { assessmentAxiosInstance } from '../assessmentServerConfig'
import { Evaluation } from '../../interfaces/evaluation'

export const getSelfEvaluationsForParticipantInPhase = async (
  coursePhaseID: string,
  courseParticipationID: string,
): Promise<Evaluation[]> => {
  const response = await assessmentAxiosInstance.get<Evaluation[]>(
    `assessment/api/course_phase/${coursePhaseID}/evaluation/self/${courseParticipationID}`,
  )
  return response.data
}
