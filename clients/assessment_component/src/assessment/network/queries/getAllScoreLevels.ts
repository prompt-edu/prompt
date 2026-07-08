import type { ScoreLevelWithParticipation } from '../../interfaces/scoreLevelWithParticipation'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllScoreLevels = async (
  coursePhaseID: string,
): Promise<ScoreLevelWithParticipation[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment/scoreLevel`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
