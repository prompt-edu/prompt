import { create } from 'zustand'
import type { Evaluation } from '../interfaces/evaluation'
import type { EvaluationCompletion } from '../interfaces/evaluationCompletion'

export interface EvaluationStore {
  selfEvaluations: Evaluation[]
  setSelfEvaluations: (evaluations: Evaluation[]) => void
  peerEvaluations: Evaluation[]
  setPeerEvaluations: (evaluations: Evaluation[]) => void
  tutorEvaluations: Evaluation[]
  setTutorEvaluations: (evaluations: Evaluation[]) => void

  selfEvaluationCompletion: EvaluationCompletion | undefined
  setSelfEvaluationCompletion: (completion: EvaluationCompletion | undefined) => void
  peerEvaluationCompletions: EvaluationCompletion[]
  setPeerEvaluationCompletions: (completions: EvaluationCompletion[]) => void
  tutorEvaluationCompletions: EvaluationCompletion[]
  setTutorEvaluationCompletions: (completions: EvaluationCompletion[]) => void
}

export const useEvaluationStore = create<EvaluationStore>((set) => ({
  selfEvaluations: [],
  setSelfEvaluations: (evaluations) => set({ selfEvaluations: evaluations }),
  peerEvaluations: [],
  setPeerEvaluations: (evaluations) => set({ peerEvaluations: evaluations }),
  tutorEvaluations: [],
  setTutorEvaluations: (evaluations) => set({ tutorEvaluations: evaluations }),

  selfEvaluationCompletion: undefined,
  setSelfEvaluationCompletion: (completion) => set({ selfEvaluationCompletion: completion }),
  peerEvaluationCompletions: [],
  setPeerEvaluationCompletions: (completions) => set({ peerEvaluationCompletions: completions }),
  tutorEvaluationCompletions: [],
  setTutorEvaluationCompletions: (completions) => set({ tutorEvaluationCompletions: completions }),
}))
