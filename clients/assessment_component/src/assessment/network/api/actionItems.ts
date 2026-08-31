import type {
  ActionItem,
  CreateActionItemRequest,
  UpdateActionItemRequest,
} from '../../interfaces/actionItem'
import { assessmentRequest, coursePhasePath } from '../client'

const path = (coursePhaseID: string) =>
  `${coursePhasePath(coursePhaseID)}/student-assessment/action-item`

export const actionItems = {
  ofParticipant: (coursePhaseID: string, courseParticipationID: string): Promise<ActionItem[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/course-participation/${courseParticipationID}`),

  listMine: (coursePhaseID: string): Promise<ActionItem[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/my-action-items`),

  create: (coursePhaseID: string, actionItem: CreateActionItemRequest): Promise<void> =>
    assessmentRequest.post(path(coursePhaseID), actionItem),

  update: (coursePhaseID: string, actionItem: UpdateActionItemRequest): Promise<void> =>
    assessmentRequest.put(`${path(coursePhaseID)}/${actionItem.id}`, actionItem),

  remove: (coursePhaseID: string, actionItemID: string): Promise<void> =>
    assessmentRequest.del(`${path(coursePhaseID)}/${actionItemID}`),
}
