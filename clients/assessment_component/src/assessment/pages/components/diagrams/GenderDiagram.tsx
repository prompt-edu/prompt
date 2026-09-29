import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  groupByGender,
} from '@tumaet/prompt-ui-components'
import { GradeDistributionBarChart } from './gradeDistributionBarChart/GradeDistributionBarChart'
import { createGradeDistributionDataPoint } from './gradeDistributionBarChart/utils/createGradeDistributionDataPoint'
import type { ParticipationWithAssessment } from './interfaces/ParticipationWithAssessment'
import { ScoreDistributionBarChart } from './scoreDistributionBarChart/ScoreDistributionBarChart'
import { createScoreDistributionDataPoint } from './scoreDistributionBarChart/utils/createScoreDistributionDataPoint'

interface GenderDiagramProps {
  participationsWithAssessment: ParticipationWithAssessment[]
  showGrade?: boolean
}

export const GenderDiagram = ({
  participationsWithAssessment,
  showGrade = false,
}: GenderDiagramProps) => {
  const data = groupByGender(participationsWithAssessment, (p) => p.participation.student)

  return (
    <Card className='flex flex-col'>
      <CardHeader className='items-center'>
        <CardTitle>Gender Distribution </CardTitle>
        <CardDescription>Scores</CardDescription>
      </CardHeader>
      <CardContent className='flex-1'>
        {showGrade ? (
          <GradeDistributionBarChart
            chartTitle='Grade distribution by gender'
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
            chartTitle='Score distribution by gender'
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
