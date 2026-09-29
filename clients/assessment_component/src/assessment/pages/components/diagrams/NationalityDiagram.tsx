import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  groupByNationality,
} from '@tumaet/prompt-ui-components'
import { GradeDistributionBarChart } from './gradeDistributionBarChart/GradeDistributionBarChart'
import { createGradeDistributionDataPoint } from './gradeDistributionBarChart/utils/createGradeDistributionDataPoint'

import type { ParticipationWithAssessment } from './interfaces/ParticipationWithAssessment'
import { ScoreDistributionBarChart } from './scoreDistributionBarChart/ScoreDistributionBarChart'
import { createScoreDistributionDataPoint } from './scoreDistributionBarChart/utils/createScoreDistributionDataPoint'

import { getGridSpanClass } from './utils/getGridSpanClass'

interface NationalityDiagramProps {
  participationsWithAssessment: ParticipationWithAssessment[]
  showGrade?: boolean
}

export const NationalityDiagram = ({
  participationsWithAssessment,
  showGrade = false,
}: NationalityDiagramProps) => {
  const data = groupByNationality(participationsWithAssessment, (p) => p.participation.student)

  return (
    <Card className={`flex flex-col ${getGridSpanClass(data.length)}`}>
      <CardHeader className='items-center pb-0'>
        <CardTitle>Nationality Distribution</CardTitle>
        <CardDescription>Scores</CardDescription>
      </CardHeader>
      <CardContent className='flex-1 pb-0'>
        {showGrade ? (
          <GradeDistributionBarChart
            chartTitle='Grade distribution by nationality'
            data={data.map((d) =>
              createGradeDistributionDataPoint(
                d.shortLabel,
                d.label,
                d.items
                  .map((p) => p.assessmentCompletion?.gradeSuggestion)
                  .filter((grade): grade is number => grade !== undefined),
              ),
            )}
          />
        ) : (
          <ScoreDistributionBarChart
            chartTitle='Score distribution by nationality'
            data={data.map((d) =>
              createScoreDistributionDataPoint(
                d.shortLabel,
                d.label,
                d.items.map((p) => p.scoreNumeric),
                d.items.map((p) => p.scoreLevel),
              ),
            )}
          />
        )}
      </CardContent>
    </Card>
  )
}
