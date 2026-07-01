import { Gender, getGenderString } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import { ApplicationParticipation } from '../../../interfaces/applicationParticipation'
import { StackedBarChartWithPassStatus } from './StackedBarChartWithPassStatus'

interface GenderDistributionCardProps {
  applications: ApplicationParticipation[]
}

export const ApplicationGenderDiagram = ({ applications }: GenderDistributionCardProps) => {
  const genderData = useMemo(() => {
    // Initialize counts for each gender
    const initialCounts: Record<Gender, { passed: number; failed: number; not_assessed: number }> =
      {
        [Gender.MALE]: { passed: 0, failed: 0, not_assessed: 0 },
        [Gender.FEMALE]: { passed: 0, failed: 0, not_assessed: 0 },
        [Gender.DIVERSE]: { passed: 0, failed: 0, not_assessed: 0 },
        [Gender.PREFER_NOT_TO_SAY]: { passed: 0, failed: 0, not_assessed: 0 },
      }

    // Aggregate counts by gender and passStatus
    const countsByGender = applications.reduce((acc, app) => {
      const gender: Gender = app.student.gender || Gender.PREFER_NOT_TO_SAY
      acc[gender][app.passStatus] += 1
      return acc
    }, initialCounts)

    // Map the counts to diagram data for the chart
    const diagramData = (Object.values(Gender) as Gender[]).map((gender) => {
      const accepted = countsByGender[gender]?.passed ?? 0
      const rejected = countsByGender[gender]?.failed ?? 0
      const notAssessed = countsByGender[gender]?.not_assessed ?? 0

      return {
        dataKey: gender !== Gender.PREFER_NOT_TO_SAY ? getGenderString(gender) : 'Unknown',
        accepted,
        rejected,
        notAssessed,
        total: accepted + rejected + notAssessed,
      }
    })

    return diagramData
  }, [applications])

  return (
    <Card className='flex flex-col w-full h-full'>
      <CardHeader className='items-center'>
        <CardTitle>Gender Distribution</CardTitle>
        <CardDescription>Breakdown of student genders</CardDescription>
      </CardHeader>
      <CardContent className='flex-1 flex flex-col justify-end pb-0'>
        <StackedBarChartWithPassStatus data={genderData} />
      </CardContent>
    </Card>
  )
}
