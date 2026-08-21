import type { ScoreLevel } from '@tumaet/prompt-shared-state'
import type { InterviewAnswer } from './InterviewAnswer'

export interface InterviewReview {
  courseParticipationID: string
  score?: number
  scoreLevel?: ScoreLevel
  interviewer: string
  interviewAnswers: InterviewAnswer[]
}

export interface UpdateInterviewReviewRequest {
  score?: number
  interviewer: string
  interviewAnswers: InterviewAnswer[]
}
