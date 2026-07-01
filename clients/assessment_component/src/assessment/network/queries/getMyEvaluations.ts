import { Evaluation } from '../../interfaces/evaluation'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getMyEvaluations = async (coursePhaseID: string): Promise<Evaluation[]> => {
  const response = await assessmentAxiosInstance.get<Evaluation[]>(
    `assessment/api/course_phase/${coursePhaseID}/evaluation/my-evaluations`,
  )
  return response.data
}
