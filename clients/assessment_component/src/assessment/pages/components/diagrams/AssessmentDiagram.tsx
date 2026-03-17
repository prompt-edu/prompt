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

import { AssessmentType } from '../../../interfaces/assessmentType'
import { AssessmentParticipationWithStudent } from '../../../interfaces/assessmentParticipationWithStudent'
import { ScoreLevelWithParticipation } from '../../../interfaces/scoreLevelWithParticipation'
import { CompetencyScoreCompletion } from '../../../interfaces/competencyScoreCompletion'

const chartConfig = {
  notAssessed: {
    label: 'Not Assessed',
    color: 'hsl(var(--muted))',
  },
  inProgress: {
    label: 'In Progress',
    color: '#63B3ED', // Light blue
  },
  completed: {
    label: 'Completed',
    color: 'hsl(var(--success))',
  },
} satisfies ChartConfig

interface AssessmentDiagramProps {
  participations: AssessmentParticipationWithStudent[]
  scoreLevels: ScoreLevelWithParticipation[]
  completions: CompetencyScoreCompletion[]
  assessmentType?: AssessmentType
}

export const AssessmentDiagram = ({
  participations,
  scoreLevels,
  completions,
  assessmentType = AssessmentType.ASSESSMENT,
}: AssessmentDiagramProps) => {
  const { chartData, totalAssessments } = React.useMemo(() => {
    const completed = participations.filter((p) =>
      completions?.find((c) => c.courseParticipationID === p.courseParticipationID && c.completed),
    ).length

    const inProgress = participations.filter(
      (p) =>
        scoreLevels.some((sl) => sl.courseParticipationID === p.courseParticipationID) &&
        !completions?.find((c) => c.courseParticipationID === p.courseParticipationID)?.completed,
    ).length

    const notAssessed = participations.length - completed - inProgress

    return {
      chartData: [
        { status: 'notAssessed', applications: notAssessed, fill: chartConfig.notAssessed.color },
        { status: 'inProgress', applications: inProgress, fill: chartConfig.inProgress.color },
        { status: 'completed', applications: completed, fill: chartConfig.completed.color },
      ],
      totalAssessments: participations.length,
    }
  }, [participations, completions, scoreLevels])

  return (
    <Card className='flex flex-col'>
      <CardHeader className='items-center pb-0'>
        <CardTitle>
          {(() => {
            switch (assessmentType) {
              case AssessmentType.SELF:
                return 'Self Evaluation'
              case AssessmentType.PEER:
                return 'Peer Evaluation'
              default:
                return 'Assessments'
            }
          })()}
        </CardTitle>
        <CardDescription>
          {(() => {
            switch (assessmentType) {
              case AssessmentType.SELF:
                return 'self evaluations '
              case AssessmentType.PEER:
                return 'peer evaluations '
              default:
                return 'assessments '
            }
          })()}
          and their status
        </CardDescription>
      </CardHeader>
      <CardContent className='flex-1 pb-0'>
        <ChartContainer config={chartConfig} className='mx-auto aspect-square max-h-[250px]'>
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
                          {totalAssessments.toLocaleString()}
                        </tspan>
                        <tspan
                          x={viewBox.cx}
                          y={(viewBox.cy || 0) + 24}
                          className='fill-muted-foreground'
                        >
                          Assessments
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
