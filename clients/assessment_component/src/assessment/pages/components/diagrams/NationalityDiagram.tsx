import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { getCountryName } from '@tumaet/prompt-shared-state'

import { ScoreDistributionBarChart } from './scoreDistributionBarChart/ScoreDistributionBarChart'
import { GradeDistributionBarChart } from './gradeDistributionBarChart/GradeDistributionBarChart'

import { ParticipationWithAssessment } from './interfaces/ParticipationWithAssessment'

import { createScoreDistributionDataPoint } from './scoreDistributionBarChart/utils/createScoreDistributionDataPoint'
import { createGradeDistributionDataPoint } from './gradeDistributionBarChart/utils/createGradeDistributionDataPoint'

import { getGridSpanClass } from './utils/getGridSpanClass'
import { groupBy } from './utils/groupBy'

interface NationalityDiagramProps {
  participationsWithAssessment: ParticipationWithAssessment[]
  showGrade?: boolean
}

export const NationalityDiagram = ({
  participationsWithAssessment,
  showGrade = false,
}: NationalityDiagramProps) => {
  const data = Array.from(
    groupBy(participationsWithAssessment, (p) => p.participation.student.nationality || 'Unknown'),
  ).map(([nationality, participations]) => {
    return {
      shortLabel: nationality,
      label: getCountryName(nationality) ?? 'Unknown',
      participationWithAssessment: participations,
    }
  })

  return (
    <Card className={`flex flex-col ${getGridSpanClass(data.length)}`}>
      <CardHeader className='items-center pb-0'>
        <CardTitle>Nationality Distribution</CardTitle>
        <CardDescription>Scores</CardDescription>
      </CardHeader>
      <CardContent className='flex-1 pb-0'>
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
