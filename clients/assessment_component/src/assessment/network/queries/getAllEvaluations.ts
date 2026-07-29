import type { Evaluation } from '../../interfaces/evaluation'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllEvaluations = async (coursePhaseID: string): Promise<Evaluation[]> => {
  const response = await assessmentAxiosInstance.get<Evaluation[]>(
    `assessment/api/course_phase/${coursePhaseID}/evaluation`,
  )
  return response.data
}
