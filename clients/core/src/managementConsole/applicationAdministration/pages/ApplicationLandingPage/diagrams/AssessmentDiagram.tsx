import * as React from 'react'
import { Label, Pie, PieChart } from 'recharts'

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

const chartConfig = {
  applications: {
    label: 'Applications',
  },
  notAssessed: {
    label: 'Not Assessed',
    color: 'hsl(var(--muted))',
  },
  accepted: {
    label: 'Accepted',
    color: 'hsl(var(--success))',
  },
  rejected: {
    label: 'Rejected',
    color: 'hsl(var(--destructive))',
  },
} satisfies ChartConfig

interface AssessmentDiagramProps {
  applications: ApplicationParticipation[]
}

export const AssessmentDiagram = ({ applications }: AssessmentDiagramProps) => {
  const { chartData, totalApplications } = React.useMemo(() => {
    const notAssessed = applications.filter((app) => app.passStatus === 'not_assessed').length
    const accepted = applications.filter((app) => app.passStatus === 'passed').length
    const rejected = applications.filter((app) => app.passStatus === 'failed').length

    return {
      chartData: [
        { status: 'notAssessed', applications: notAssessed, fill: chartConfig.notAssessed.color },
        { status: 'accepted', applications: accepted, fill: chartConfig.accepted.color },
        { status: 'rejected', applications: rejected, fill: chartConfig.rejected.color },
      ],
      totalApplications: applications.length,
    }
  }, [applications])

  return (
    <Card className='flex flex-col'>
      <CardHeader className='items-center pb-0'>
        <CardTitle>Applications</CardTitle>
        <CardDescription>All applications and their assessment status</CardDescription>
      </CardHeader>
      <CardContent className='flex-1 pb-0 min-h-[250px]'>
        <ChartContainer
          config={chartConfig}
          className='mx-auto w-full h-[250px] min-h-[250px] min-w-0'
        >
          <PieChart>
            <ChartTooltip cursor={false} content={<ChartTooltipContent hideLabel />} />
            <Pie
              data={chartData}
              dataKey='applications'
              nameKey='status'
              innerRadius={60}
              strokeWidth={5}
            >
              <Label
                content={({ viewBox }) => {
                  if (viewBox && 'cx' in viewBox && 'cy' in viewBox) {
                    return (
                      <text
                        x={viewBox.cx}
                        y={viewBox.cy}
                        textAnchor='middle'
                        dominantBaseline='middle'
                      >
                        <tspan
                          x={viewBox.cx}
                          y={viewBox.cy}
                          className='fill-foreground text-3xl font-bold'
                        >
                          {totalApplications.toLocaleString()}
                        </tspan>
                        <tspan
                          x={viewBox.cx}
                          y={(viewBox.cy || 0) + 24}
                          className='fill-muted-foreground'
                        >
                          Applications
                        </tspan>
                      </text>
                    )
                  }
                }}
              />
            </Pie>
          </PieChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}
