import { Competency, UpdateCompetencyRequest } from '../../interfaces/competency'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const updateCompetency = async (
  coursePhaseID: string,
  competency: UpdateCompetencyRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.put<Competency>(
      `assessment/api/course_phase/${coursePhaseID}/competency/${competency.id}`,
      competency,
      {
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
