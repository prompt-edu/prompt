import type { ActionItem, CreateActionItemRequest } from '../../interfaces/actionItem'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const createActionItem = async (
  coursePhaseID: string,
  actionItem: CreateActionItemRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.post<ActionItem>(
      `assessment/api/course_phase/${coursePhaseID}/student-assessment/action-item`,
      actionItem,
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
