import { useStudyPrograms } from '@core/network/hooks/useStudyPrograms'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Skeleton,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import type { ApplicationParticipation } from '../../../interfaces/applicationParticipation'
import { StackedBarChartWithPassStatus } from './StackedBarChartWithPassStatus'
import { groupApplicationsByStudyProgram } from './utils/groupApplicationsByStudyProgram'

interface StudyBackgroundCardProps {
  applications: ApplicationParticipation[]
}

export const ApplicationStudyBackgroundDiagram = ({ applications }: StudyBackgroundCardProps) => {
  const { data: studyPrograms } = useStudyPrograms()

  const studyData = useMemo(
    () => studyPrograms && groupApplicationsByStudyProgram(applications, studyPrograms),
    [applications, studyPrograms],
  )

  return (
    <Card className='flex flex-col w-full h-full'>
      <CardHeader className='items-center'>
        <CardTitle>Study Program Distribution</CardTitle>
        <CardDescription>Breakdown of student study programs</CardDescription>
      </CardHeader>
      <CardContent className='flex-1 flex flex-col justify-end pb-0'>
        {studyData ? (
          <StackedBarChartWithPassStatus data={studyData} />
        ) : (
          <Skeleton className='w-full h-[280px]' />
        )}
      </CardContent>
    </Card>
  )
}
