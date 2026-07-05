import { create } from 'zustand'
import type { Assessment } from '../interfaces/assessment'
import type { AssessmentCompletion } from '../interfaces/assessmentCompletion'
import type { AssessmentParticipationWithStudent } from '../interfaces/assessmentParticipationWithStudent'
import { AssessmentType } from '../interfaces/assessmentType'
import type { CategoryAssessment } from '../interfaces/categoryAssessment'
import type { Evaluation } from '../interfaces/evaluation'
import type { StudentAssessment } from '../interfaces/studentAssessment'
import type { StudentScore } from '../interfaces/studentScore'

export interface StudentAssessmentStore {
  courseParticipationID: string | undefined
  assessments: Assessment[]
  categoryAssessments: CategoryAssessment[]
  assessmentCompletion: AssessmentCompletion | undefined
  studentScore: StudentScore | undefined
  selfEvaluations: Evaluation[]
  peerEvaluations: Evaluation[]
  setStudentAssessment: (assessment: StudentAssessment) => void

  assessmentParticipation: AssessmentParticipationWithStudent | undefined
  setAssessmentParticipation: (participation: AssessmentParticipationWithStudent) => void
}

export const useStudentAssessmentStore = create<StudentAssessmentStore>((set) => ({
  courseParticipationID: undefined,
  assessments: [],
  categoryAssessments: [],
  assessmentCompletion: undefined,
  studentScore: undefined,
  selfEvaluations: [],
  peerEvaluations: [],
  setStudentAssessment: (assessment) =>
    set({
      courseParticipationID: assessment.courseParticipationID,
      assessments: assessment.assessments,
      categoryAssessments: assessment.categoryAssessments ?? [],
      assessmentCompletion: assessment.assessmentCompletion,
      studentScore: assessment.studentScore,
      selfEvaluations: assessment.evaluations.filter(
        (evaluation) => evaluation.type === AssessmentType.SELF,
      ),
      peerEvaluations: assessment.evaluations.filter(
        (evaluation) => evaluation.type === AssessmentType.PEER,
      ),
    }),

  assessmentParticipation: undefined,
  setAssessmentParticipation: (participation) => set({ assessmentParticipation: participation }),
}))
