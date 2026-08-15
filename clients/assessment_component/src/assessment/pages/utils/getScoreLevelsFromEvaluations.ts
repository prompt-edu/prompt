import { mapNumberToScoreLevel, mapScoreLevelToNumber } from '@tumaet/prompt-shared-state'
import type { AssessmentType } from '../../interfaces/assessmentType'
import type { Evaluation } from '../../interfaces/evaluation'
import type { ScoreLevelWithParticipation } from '../../interfaces/scoreLevelWithParticipation'

export const getScoreLevelsFromEvaluations = (
  evaluations: Evaluation[],
  type: AssessmentType,
): ScoreLevelWithParticipation[] => {
  const scoresByParticipant = new Map<string, number[]>()

  for (const evaluation of evaluations) {
    if (evaluation.type !== type) continue
    const scores = scoresByParticipant.get(evaluation.courseParticipationID) ?? []
    scores.push(mapScoreLevelToNumber(evaluation.scoreLevel))
    scoresByParticipant.set(evaluation.courseParticipationID, scores)
  }

  return Array.from(scoresByParticipant.entries()).map(([courseParticipationID, scores]) => {
    const scoreNumeric = scores.reduce((sum, score) => sum + score, 0) / scores.length
    return {
      courseParticipationID,
      scoreLevel: mapNumberToScoreLevel(scoreNumeric),
      scoreNumeric,
    }
  })
}
