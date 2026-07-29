import { mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import type { Assessment } from '../../../interfaces/assessment'
import type { CategoryWithCompetencies } from '../../../interfaces/category'

import { getWeightedScoreLevel } from '../../utils/getWeightedScoreLevel'

import { ScoreDistributionBarChart } from './scoreDistributionBarChart/ScoreDistributionBarChart'
import { createScoreDistributionDataPoint } from './scoreDistributionBarChart/utils/createScoreDistributionDataPoint'

import { getGridSpanClass } from './utils/getGridSpanClass'
import { groupBy } from './utils/groupBy'

interface CategoryDiagramProps {
  categories: CategoryWithCompetencies[]
  assessments: Assessment[]
}

export const CategoryDiagram = ({ categories, assessments }: CategoryDiagramProps) => {
  const chartData = categories.map((category) => {
    const categoryAssessments = assessments.filter((assessment) =>
      category.competencies.map((competency) => competency.id).includes(assessment.competencyID),
    )

    const scores = Array.from(
      groupBy(categoryAssessments, (assessment) => assessment.courseParticipationID).values(),
    ).map((participantAssessments) => getWeightedScoreLevel(participantAssessments, [category]))

    return createScoreDistributionDataPoint(
      category.shortName,
      category.name,
      scores,
      scores.map((s) => mapNumberToScoreLevel(s)),
    )
  })

  return (
    <Card className={`flex flex-col ${getGridSpanClass(categories.length)}`}>
      <CardHeader className='items-center'>
        <CardTitle>Category Distribution</CardTitle>
        <CardDescription>Scores</CardDescription>
      </CardHeader>
      <CardContent className='flex-1'>
        <ScoreDistributionBarChart data={chartData} chartTitle='Category score distribution' />
      </CardContent>
    </Card>
  )
}
