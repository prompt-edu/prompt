import type { ActionItem } from '../../interfaces/actionItem'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllActionItemsForStudentInPhase = async (
  coursePhaseID: string,
  courseParticipationID: string,
): Promise<ActionItem[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment/action-item/course-participation/${courseParticipationID}`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
