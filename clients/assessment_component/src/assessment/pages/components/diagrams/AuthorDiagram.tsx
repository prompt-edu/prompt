import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { GradeDistributionBarChart } from './gradeDistributionBarChart/GradeDistributionBarChart'
import { createGradeDistributionDataPoint } from './gradeDistributionBarChart/utils/createGradeDistributionDataPoint'
import { ParticipationWithAssessment } from './interfaces/ParticipationWithAssessment'
import { ScoreDistributionBarChart } from './scoreDistributionBarChart/ScoreDistributionBarChart'
import { createScoreDistributionDataPoint } from './scoreDistributionBarChart/utils/createScoreDistributionDataPoint'
import { getGridSpanClass } from './utils/getGridSpanClass'

import { groupBy } from './utils/groupBy'

interface AuthorDiagramProps {
  participationsWithAssessment: ParticipationWithAssessment[]
  showGrade?: boolean
}

export const AuthorDiagram = ({
  participationsWithAssessment,
  showGrade = false,
}: AuthorDiagramProps) => {
  const data = Array.from(
    groupBy(
      participationsWithAssessment,
      (p) => p.assessmentCompletion?.author ?? 'Unknown Author',
    ),
  ).map(([author, participations]) => {
    return {
      shortLabel: author
        .split(' ')
        .map((name) => name[0])
        .join(''),
      label: author,
      participationWithAssessment: participations,
    }
  })

  return (
    <Card className={`flex flex-col ${getGridSpanClass(data.length)}`}>
      <CardHeader className='items-center'>
        <CardTitle>Author Distribution</CardTitle>
        <CardDescription>Scores</CardDescription>
      </CardHeader>
      <CardContent className='flex-1'>
        {showGrade ? (
          <GradeDistributionBarChart
            data={data.map((d) =>
              createGradeDistributionDataPoint(
                d.shortLabel,
                d.label,
                d.participationWithAssessment
                  .map((p) => p.assessmentCompletion?.gradeSuggestion)
                  .filter((grade): grade is number => grade !== undefined),
              ),
            )}
          />
        ) : (
          <ScoreDistributionBarChart
            data={data.map((d) =>
              createScoreDistributionDataPoint(
                d.shortLabel,
                d.label,
                d.participationWithAssessment.map((p) => p.scoreNumeric),
                d.participationWithAssessment.map((p) => p.scoreLevel),
              ),
            )}
          />
        )}
      </CardContent>
    </Card>
  )
}
