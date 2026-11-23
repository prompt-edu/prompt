import type { PromptTableColumnDef } from '@tumaet/prompt-ui-components'
import type { ScoreLevelWithParticipation } from '../../interfaces/scoreLevelWithParticipation'
import { GradeSuggestionBadge, StudentScoreBadge } from './badges'

// The average of the evaluations someone received, shown as a score level and as a grade.

export const ScoreBadges = ({ score }: { score: ScoreLevelWithParticipation | undefined }) => (
  <>
    <StudentScoreBadge scoreLevel={score?.scoreLevel} />
    <GradeSuggestionBadge gradeSuggestion={score?.scoreNumeric} />
  </>
)

/**
 * Score Level and Grade columns. `getPersonId` returns the evaluated person's id of a row,
 * by default the row `id`.
 */
export const createScoreColumns = <T extends { id: string }>(
  scores: ScoreLevelWithParticipation[],
  getPersonId: (row: T) => string = (row) => row.id,
): PromptTableColumnDef<T>[] => {
  const scoreById = new Map(scores.map((score) => [score.courseParticipationID, score]))
  const getScore = (row: T) => scoreById.get(getPersonId(row))
  return [
    {
      id: 'scoreLevel',
      header: 'Score Level',
      accessorFn: (row) => getScore(row)?.scoreNumeric,
      cell: ({ row }) => <StudentScoreBadge scoreLevel={getScore(row.original)?.scoreLevel} />,
    },
    {
      id: 'grade',
      header: 'Grade',
      accessorFn: (row) => getScore(row)?.scoreNumeric,
      cell: ({ row }) => (
        <GradeSuggestionBadge gradeSuggestion={getScore(row.original)?.scoreNumeric} />
      ),
    },
  ]
}
