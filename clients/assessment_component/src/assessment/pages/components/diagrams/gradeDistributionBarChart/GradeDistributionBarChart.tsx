import { ChartContainer } from '@tumaet/prompt-ui-components'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Tooltip as ChartTooltip,
  Label,
  LabelList,
  XAxis,
  YAxis,
} from 'recharts'
import { GRADE_CONFIG } from '../../../utils/gradeConfig'
import { GradeDistributionBar } from './components/GradeDistributionBar'
import { GradeDistributionLabel } from './components/GradeDistributionLabel'
import { GradeDistributionTooltipContent } from './components/GradeDistributionTooltipContent'
import { GradeDistributionDataPoint } from './interfaces/GradeDistributionDataPoint'

export interface GradeDistributionBarChartProps {
  data: GradeDistributionDataPoint[]
}

export function GradeDistributionBarChart({ data }: GradeDistributionBarChartProps) {
  if (!data || data.length === 0) {
    return (
      <ChartContainer config={GRADE_CONFIG} className='w-full h-[280px]'>
        <div className='flex items-center justify-center h-64 text-gray-500'>
          <p>No data available</p>
        </div>
      </ChartContainer>
    )
  }

  const chartData = data.map((item) => ({
    shortName: item.shortLabel,
    name: item.label,
    value: 5, // Use the grade range as the value for sizing
    lowerQuartile: item.lowerQuartile,
    median: item.median,
    upperQuartile: item.upperQuartile,
    counts: item.counts,
    average: item.average,
  }))

  return (
    <ChartContainer config={GRADE_CONFIG} className='w-full h-[280px]'>
      <BarChart data={chartData} margin={{ top: 30, right: 10, bottom: 10, left: 10 }}>
        <CartesianGrid
          horizontal={true}
          vertical={false}
          strokeDasharray='5 5'
          stroke='#e5e7eb'
          opacity={1}
        />

        <XAxis dataKey='shortName' axisLine={false} tickLine={false} interval={0} />
        <YAxis
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12, fill: '#a3a3a3' }}
          domain={[1, 5]}
          ticks={[1, 2, 3, 4, 5]}
        >
          <Label value='Grade' angle={-90} position='insideLeft' fill='#a3a3a3' />
        </YAxis>
        <ChartTooltip cursor={false} content={<GradeDistributionTooltipContent />} />
        <Bar dataKey='value' shape={<GradeDistributionBar />}>
          <LabelList dataKey='average' content={<GradeDistributionLabel />} />
        </Bar>
      </BarChart>
    </ChartContainer>
  )
}
