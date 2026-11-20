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

/** Score Level and Grade columns for a table whose row `id` is the evaluated person's id. */
export const createScoreColumns = <T extends { id: string }>(
  scores: ScoreLevelWithParticipation[],
): PromptTableColumnDef<T>[] => {
  const scoreById = new Map(scores.map((score) => [score.courseParticipationID, score]))
  return [
    {
      id: 'scoreLevel',
      header: 'Score Level',
      accessorFn: (row) => scoreById.get(row.id)?.scoreNumeric,
      cell: ({ row }) => (
        <StudentScoreBadge scoreLevel={scoreById.get(row.original.id)?.scoreLevel} />
      ),
    },
    {
      id: 'grade',
      header: 'Grade',
      accessorFn: (row) => scoreById.get(row.id)?.scoreNumeric,
      cell: ({ row }) => (
        <GradeSuggestionBadge gradeSuggestion={scoreById.get(row.original.id)?.scoreNumeric} />
      ),
    },
  ]
}
