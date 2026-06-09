export interface CategoryAssessment {
  id: string
  categoryID: string
  coursePhaseID: string
  courseParticipationID: string
  comment: string
  author: string
  authorID: string
  createdAt: string
  updatedAt: string
}

export interface CreateOrUpdateCategoryAssessmentRequest {
  categoryID: string
  coursePhaseID: string
  courseParticipationID: string
  comment: string
}
