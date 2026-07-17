import { ScoreLevel } from '@tumaet/prompt-shared-state'

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import type { ScoreLevelWithParticipation } from '../../../interfaces/scoreLevelWithParticipation'

import { BarChartWithScoreLevel } from './BarChartWithScoreLevel'

interface ScoreLevelDistributionDiagramProps {
  participations: Array<{ courseParticipationID: string }>
  scoreLevels: ScoreLevelWithParticipation[]
  title?: string
  description?: string
}

export const ScoreLevelDistributionDiagram = ({
  participations,
  scoreLevels,
  title = 'Assessment Distribution',
  description = 'Number of students per score level',
}: ScoreLevelDistributionDiagramProps) => {
  const chartData = useMemo(() => {
    const scoreLevelCounts = [
      { level: ScoreLevel.VeryGood, label: 'Very Good' },
      { level: ScoreLevel.Good, label: 'Good' },
      { level: ScoreLevel.Ok, label: 'Ok' },
      { level: ScoreLevel.Bad, label: 'Bad' },
      { level: ScoreLevel.VeryBad, label: 'Very Bad' },
    ]

    // Create individual data points for each score level
    const scoreLevelData = scoreLevelCounts.map(({ level, label }) => {
      const count = scoreLevels.filter((s) => s.scoreLevel === level).length
      return {
        dataKey: label,
        count: count,
        total: participations.length,
      }
    })

    // Add not assessed data point
    const notAssessed = participations.length - scoreLevels.length
    if (notAssessed > 0) {
      scoreLevelData.push({
        dataKey: 'Not Assessed',
        count: notAssessed,
        total: participations.length,
      })
    }

    return scoreLevelData
  }, [participations, scoreLevels])

  return (
    <Card className='flex flex-col w-full h-full'>
      <CardHeader className='items-center'>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className='flex-1 flex flex-col justify-end pb-0'>
        <BarChartWithScoreLevel data={chartData} />
      </CardContent>
    </Card>
  )
}
