import { mapScoreLevelToNumber } from '@tumaet/prompt-shared-state'
import { CategoryWithCompetencies } from '../../interfaces/category'
import { CompetencyScore } from '../../interfaces/competencyScore'

export function getWeightedScoreLevel(
  competencyScores: CompetencyScore[],
  categories: CategoryWithCompetencies[],
): number {
  if (!competencyScores?.length || !categories?.length) {
    return 0
  }

  const totalWeightsOfCategoriesWithScore = categories
    .filter((category) =>
      category.competencies.some((competency) =>
        competencyScores.some((score) => score.competencyID === competency.id),
      ),
    )
    .reduce((total, category) => total + category.weight, 0)

  return (
    categories
      .map((category) => {
        const totalWeightOfCompetenciesWithScore = category.competencies
          .filter((competency) =>
            competencyScores.some((score) => score.competencyID === competency.id),
          )
          .reduce((totalWeight, competency) => totalWeight + competency.weight, 0)

        const categoryAverage = category.competencies
          .map((competency) => {
            const competencyScoresInThisCompetency = competencyScores.filter(
              (score) => score.competencyID === competency.id,
            )

            if (competencyScoresInThisCompetency.length === 0) {
              return undefined
            }

            return (
              competencyScoresInThisCompetency
                .map((score) => mapScoreLevelToNumber(score.scoreLevel) * competency.weight)
                .reduce((totalScore, score) => totalScore + score, 0) /
              competencyScoresInThisCompetency.length /
              totalWeightOfCompetenciesWithScore
            )
          })
          .reduce((totalScore: number, score) => totalScore + (score || 0), 0)

        return categoryAverage * category.weight
      })
      .reduce((totalScore, score) => totalScore + score, 0) / totalWeightsOfCategoriesWithScore
  )
}
