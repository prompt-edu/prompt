import type { ActionItem } from './actionItem'
import type { StudentAssessment } from './studentAssessment'

export interface AssessmentExport {
  exportedAt: string
  coursePhaseID: string
  courseParticipationID: string
  studentAssessment: StudentAssessment
  actionItems: ActionItem[]
}
