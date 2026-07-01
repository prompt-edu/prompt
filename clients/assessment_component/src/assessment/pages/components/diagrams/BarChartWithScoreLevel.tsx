import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@tumaet/prompt-ui-components'
import { Bar, BarChart, Cell, LabelList, XAxis, YAxis } from 'recharts'
import { chartConfig } from './utils/chartConfig'

interface ScoreLevelDataPoint {
  dataKey: string
  count: number
  total: number
}

interface BarChartWithScoreLevelProps {
  data: ScoreLevelDataPoint[]
}

// Score level color mapping
const getColorForScoreLevel = (scoreLevel: string): string => {
  switch (scoreLevel) {
    case 'Very Good':
      return chartConfig.veryGood?.color || '#93c5fd'
    case 'Good':
      return chartConfig.good?.color || '#86efac'
    case 'Ok':
      return chartConfig.ok?.color || '#fde68a'
    case 'Bad':
      return chartConfig.bad?.color || '#f97316'
    case 'Very Bad':
      return chartConfig.veryBad?.color || '#fca5a5'
    case 'Not Assessed':
      return chartConfig.notAssessed?.color || '#d4d4d8'
    default:
      return '#d4d4d8'
  }
}

export const BarChartWithScoreLevel = ({ data }: BarChartWithScoreLevelProps) => {
  return (
    <ChartContainer config={chartConfig} className='mx-auto w-full h-[280px]'>
      <BarChart data={data} margin={{ top: 20, right: 20, bottom: 20, left: 20 }}>
        <XAxis
          dataKey='dataKey'
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 11 }}
          interval={0}
          height={60}
          angle={-45}
          textAnchor='end'
        />
        <YAxis hide />
        <ChartTooltip
          cursor={false}
          content={<ChartTooltipContent />}
          labelFormatter={(label) => `Score Level: ${label}`}
          formatter={(value) => [value, 'Students']}
        />
        <Bar dataKey='count' radius={[4, 4, 0, 0]}>
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={getColorForScoreLevel(entry.dataKey)} />
          ))}
          <LabelList
            dataKey='count'
            position='top'
            offset={10}
            className='fill-foreground'
            fontSize={12}
          />
        </Bar>
      </BarChart>
    </ChartContainer>
  )
}
