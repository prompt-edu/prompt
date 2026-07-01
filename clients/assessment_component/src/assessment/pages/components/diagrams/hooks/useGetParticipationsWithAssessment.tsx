import { useMemo } from 'react'
import type { Assessment } from '../../../../interfaces/assessment'
import type { AssessmentCompletion } from '../../../../interfaces/assessmentCompletion'
import type { AssessmentParticipationWithStudent } from '../../../../interfaces/assessmentParticipationWithStudent'
import type { ScoreLevelWithParticipation } from '../../../../interfaces/scoreLevelWithParticipation'

import type { ParticipationWithAssessment } from '../interfaces/ParticipationWithAssessment'

export const useGetParticipationsWithAssessment = (
  participations: AssessmentParticipationWithStudent[],
  participationsWithScore: ScoreLevelWithParticipation[],
  assessmentCompletions: AssessmentCompletion[],
  assessments: Assessment[],
) => {
  return useMemo<ParticipationWithAssessment[]>(() => {
    const temp = participationsWithScore
      .map((participationWithScore) => {
        const participation = participations.find(
          (p) => p.courseParticipationID === participationWithScore.courseParticipationID,
        )
        const completion = assessmentCompletions.find(
          (c) => c.courseParticipationID === participationWithScore.courseParticipationID,
        )

        return {
          participation,
          assessments: assessments.filter(
            (a) => a.courseParticipationID === participationWithScore.courseParticipationID,
          ),
          scoreLevel: participationWithScore.scoreLevel,
          scoreNumeric: participationWithScore.scoreNumeric,
          assessmentCompletion: completion,
        } as ParticipationWithAssessment
      })
      .filter((p) => p.participation !== undefined)
    return temp
  }, [participations, participationsWithScore, assessmentCompletions, assessments])
}
