import { mapScoreLevelToNumber, ScoreLevel } from '@tumaet/prompt-shared-state'
import { StudentScoreBadge } from '../../components/badges'
import { getLevelConfig } from '../../utils/getLevelConfig'
import { ScoreLevelWithParticipation } from '../../../interfaces/scoreLevelWithParticipation'
import {
  ExtraParticipantColumn,
  ParticipantRow,
} from '@/components/pages/CoursePhaseParticipationsTable/table/participationRow'

export const createScoreLevelColumn = (
  scoreLevels: ScoreLevelWithParticipation[],
): ExtraParticipantColumn<ScoreLevel | null> => {
  const getScoreLevelForParticipation = (courseParticipationID: string): ScoreLevel | null => {
    const match = scoreLevels.find((s) => s.courseParticipationID === courseParticipationID)
    return match ? match.scoreLevel : null
  }

  return {
    id: 'scoreLevel',
    header: 'Score Level',

    accessorFn: (row: ParticipantRow) => getScoreLevelForParticipation(row.courseParticipationID),

    cell: ({ getValue }) => {
      const scoreLevel = getValue()
      return scoreLevel ? <StudentScoreBadge scoreLevel={scoreLevel} /> : null
    },

    enableSorting: true,
    sortingFn: (rowA, rowB) => {
      const a =
        getScoreLevelForParticipation(rowA.original.courseParticipationID) ?? ScoreLevel.VeryBad
      const b =
        getScoreLevelForParticipation(rowB.original.courseParticipationID) ?? ScoreLevel.VeryBad

      return mapScoreLevelToNumber(a) - mapScoreLevelToNumber(b)
    },

    enableColumnFilter: true,

    extraData: scoreLevels.map((s) => ({
      courseParticipationID: s.courseParticipationID,
      value: s.scoreLevel,
      stringValue: getLevelConfig(s.scoreLevel).title,
    })),
  }
}
