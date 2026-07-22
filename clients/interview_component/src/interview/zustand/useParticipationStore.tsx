import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { create } from 'zustand'
import type { InterviewReview } from '../interfaces/InterviewReview'
import type { InterviewSlot } from '../interfaces/InterviewSlots'

export interface ParticipationStore {
  participations: CoursePhaseParticipationWithStudent[]
  interviewSlots: InterviewSlot[]
  interviewReviews: Record<string, InterviewReview>
  setParticipations: (participations: CoursePhaseParticipationWithStudent[]) => void
  setInterviewSlots: (interviewSlots: InterviewSlot[]) => void
  setInterviewReviews: (interviewReviews: InterviewReview[]) => void
}

export const useParticipationStore = create<ParticipationStore>((set) => ({
  participations: [],
  interviewSlots: [],
  interviewReviews: {},
  setParticipations: (participations) => set({ participations }),
  setInterviewSlots: (interviewSlots) => set({ interviewSlots }),
  setInterviewReviews: (interviewReviews) =>
    set({
      interviewReviews: Object.fromEntries(
        interviewReviews.map((review) => [review.courseParticipationID, review]),
      ),
    }),
}))
