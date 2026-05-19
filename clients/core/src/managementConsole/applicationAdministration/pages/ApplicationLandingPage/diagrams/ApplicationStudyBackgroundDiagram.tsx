import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { ApplicationParticipation } from '../../../interfaces/applicationParticipation'
import { useMemo } from 'react'
import { translations } from '@tumaet/prompt-shared-state'
import { StackedBarChartWithPassStatus } from './StackedBarChartWithPassStatus'

const programsWithOther = translations.university.studyPrograms.concat('Other')

interface StudyBackgroundCardProps {
  applications: ApplicationParticipation[]
}

export const ApplicationStudyBackgroundDiagram = ({ applications }: StudyBackgroundCardProps) => {
  const studyData = useMemo(() => {
    // Use the complete list of programs including 'Other'
    const allPrograms = programsWithOther

    // Count each pass status per program, initializing the count object when needed.
    const countsByProgram = applications.reduce(
      (acc, app) => {
        const program = app.student.studyProgram || 'Other'
        if (!acc[program]) {
          acc[program] = { passed: 0, failed: 0, not_assessed: 0 }
        }
        acc[program][app.passStatus] = (acc[program][app.passStatus] || 0) + 1
        return acc
      },
      {} as Record<string, { passed: number; failed: number; not_assessed: number }>,
    )

    // Map over all programs (including "Other") to create the data for the chart.
    const data = allPrograms.map((program) => {
      const accepted = countsByProgram[program]?.passed ?? 0
      const rejected = countsByProgram[program]?.failed ?? 0
      const notAssessed = countsByProgram[program]?.not_assessed ?? 0

      return {
        program,
        dataKey: translations.university.studyProgramShortNames[program] || program,
        accepted,
        rejected,
        notAssessed,
        total: accepted + rejected + notAssessed,
      }
    })

    return data
  }, [applications])

  return (
    <Card className='flex flex-col w-full h-full'>
      <CardHeader className='items-center'>
        <CardTitle>Study Program Distribution</CardTitle>
        <CardDescription>Breakdown of student study programs</CardDescription>
      </CardHeader>
      <CardContent className='flex-1 flex flex-col justify-end pb-0'>
        <StackedBarChartWithPassStatus data={studyData} />
      </CardContent>
    </Card>
  )
}
