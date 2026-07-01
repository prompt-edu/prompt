import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'

import { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import { VALID_GRADE_VALUES } from '../../utils/gradeConfig'

import { BarChartWithGrades } from './BarChartWithGrades'

interface GradeDistributionDiagramProps {
  participations: AssessmentParticipationWithStudent[]
  grades: number[]
}

export const GradeDistributionDiagram = ({
  participations,
  grades,
}: GradeDistributionDiagramProps) => {
  const chartData = useMemo(() => {
    const gradeData = VALID_GRADE_VALUES.map((gradeValue) => {
      const count = grades.filter((grade) => Math.abs(grade - gradeValue) < 0.01).length

      return {
        dataKey: gradeValue.toFixed(1),
        count: count,
        total: participations.length,
      }
    })

    // Add not graded data point
    const notGraded = participations.length - grades.length
    if (notGraded > 0) {
      gradeData.push({
        dataKey: 'No Grade',
        count: notGraded,
        total: participations.length,
      })
    }

    return gradeData
  }, [participations, grades])

  const averageGrade = useMemo(() => {
    if (grades.length === 0) return null
    const total = grades.reduce((sum, grade) => sum + grade, 0)
    return (total / grades.length).toFixed(1)
  }, [grades])

  return (
    <Card className='flex flex-col w-full h-full'>
      <CardHeader className='items-center'>
        <CardTitle>Grade Distribution</CardTitle>
        <CardDescription>
          Number of students per grade {averageGrade && ` (Average: ${averageGrade})`}
        </CardDescription>
      </CardHeader>
      <CardContent className='flex-1 flex flex-col justify-end pb-0'>
        <BarChartWithGrades data={chartData} />
      </CardContent>
    </Card>
  )
}
