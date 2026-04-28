import { Gender } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'

import { ParticipationWithAssessment } from './interfaces/ParticipationWithAssessment'

import { ScoreDistributionBarChart } from './scoreDistributionBarChart/ScoreDistributionBarChart'
import { GradeDistributionBarChart } from './gradeDistributionBarChart/GradeDistributionBarChart'

import { createScoreDistributionDataPoint } from './scoreDistributionBarChart/utils/createScoreDistributionDataPoint'
import { createGradeDistributionDataPoint } from './gradeDistributionBarChart/utils/createGradeDistributionDataPoint'

interface GenderDiagramProps {
  participationsWithAssessment: ParticipationWithAssessment[]
  showGrade?: boolean
}

export const GenderDiagram = ({
  participationsWithAssessment,
  showGrade = false,
}: GenderDiagramProps) => {
  const data = (Object.values(Gender) as Gender[]).map((gender) => {
    const genderLabel =
      gender === Gender.PREFER_NOT_TO_SAY
        ? 'Unknown'
        : gender.replace(/_/g, ' ').charAt(0).toUpperCase() + gender.replace(/_/g, ' ').slice(1)

    const participationWithAssessment = participationsWithAssessment.filter(
      (p) => p.participation.student.gender === gender,
    )

    return {
      shortLabel: genderLabel,
      label: genderLabel,
      participationWithAssessment: participationWithAssessment,
    }
  })

  return (
    <Card className='flex flex-col'>
      <CardHeader className='items-center'>
        <CardTitle>Gender Distribution </CardTitle>
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
