import { Bar, BarChart, XAxis, YAxis } from 'recharts'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@tumaet/prompt-ui-components'
import { ApplicationParticipation } from '../../../interfaces/applicationParticipation'
import { useMemo } from 'react'

const chartConfig = {
  bachelor: {
    label: 'Bachelor',
    color: 'hsl(var(--primary))',
  },
  master: {
    label: 'Master',
    color: 'hsl(var(--chart-blue))',
  },
} satisfies ChartConfig

interface ApplicationStudySemesterDiagramProps {
  applications: ApplicationParticipation[]
}

export const ApplicationStudySemesterDiagram = ({
  applications,
}: ApplicationStudySemesterDiagramProps) => {
  const { semesterData } = useMemo(() => {
    const semesterCounts = applications.reduce(
      (acc, app) => {
        const semester = app.student.currentSemester || 1
        const degree = app.student.studyDegree || 'bachelor'
        const key = semester > 12 ? '12+' : semester.toString()

        acc[key] = acc[key] || { bachelor: 0, master: 0 }
        acc[key][degree] = (acc[key][degree] || 0) + 1
        return acc
      },
      {} as Record<string, { bachelor: number; master: number }>,
    )

    const data = Object.entries(semesterCounts).map(([semester, counts]) => ({
      semester,
      Bachelor: counts.bachelor,
      Master: counts.master,
    }))

    return { semesterData: data }
  }, [applications])

  return (
    <Card className='flex flex-col w-full h-full'>
      <CardHeader className='items-center'>
        <CardTitle>Semester Distribution</CardTitle>
        <CardDescription>Breakdown of students by semester and degree</CardDescription>
      </CardHeader>
      <CardContent className='flex-1 flex flex-col justify-end pb-0 min-h-[280px]'>
        <ChartContainer
          config={chartConfig}
          className='mx-auto w-full h-[280px] min-h-[280px] min-w-0'
        >
          <BarChart data={semesterData} margin={{ top: 30, right: 10, bottom: 0, left: 10 }}>
            <XAxis
              dataKey='semester'
              axisLine={false}
              tickLine={false}
              tick={{ fontSize: 12 }}
              interval={0}
              height={50}
            />
            <YAxis hide />
            <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
            <Bar dataKey='Bachelor' radius={[4, 4, 0, 0]} fill={chartConfig.bachelor.color} />
            <Bar dataKey='Master' radius={[4, 4, 0, 0]} fill={chartConfig.master.color} />
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}
