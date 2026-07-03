import type { ActionItem } from '../../interfaces/actionItem'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getMyActionItems = async (coursePhaseID: string): Promise<ActionItem[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment/action-item/my-action-items`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
