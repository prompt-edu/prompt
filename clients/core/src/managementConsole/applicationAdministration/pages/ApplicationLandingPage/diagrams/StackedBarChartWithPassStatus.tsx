import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@tumaet/prompt-ui-components'
import { BarChart, Bar, LabelList, XAxis, YAxis, Rectangle } from 'recharts'
import { chartConfig } from './utils/chartConfig'
import { getCornerRadius } from './utils/getCornerRadius'
import { DataPoint } from '../interfaces/DataPoint'

interface StackedBarChartWithPassStatusProps {
  data: DataPoint[]
}

const createRoundedBarShape = (segmentKey: string) => {
  const RoundedBarShape = (props: any) => {
    const { x, y, width, height, payload } = props
    // Compute the corner radius for the current segment
    const radius = getCornerRadius(payload, segmentKey)
    return (
      <Rectangle
        x={x}
        y={y}
        width={width}
        height={height}
        radius={radius}
        fill={chartConfig[segmentKey]?.color}
      />
    )
  }
  RoundedBarShape.displayName = `RoundedBarShape(${segmentKey})`
  return RoundedBarShape
}

export const StackedBarChartWithPassStatus = ({ data }: StackedBarChartWithPassStatusProps) => {
  return (
    <ChartContainer config={chartConfig} className='mx-auto w-full h-[280px] min-h-[280px] min-w-0'>
      <BarChart data={data} margin={{ top: 30, right: 10, bottom: 0, left: 10 }}>
        <XAxis
          dataKey={'dataKey'}
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12 }}
          interval={0}
          height={50}
          tickFormatter={(value) => value.toLocaleString()}
        />
        <YAxis hide />
        <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
        <Bar dataKey='accepted' stackId='passStatus' shape={createRoundedBarShape('accepted')} />
        <Bar dataKey='rejected' stackId='passStatus' shape={createRoundedBarShape('rejected')} />
        <Bar
          dataKey='notAssessed'
          stackId='passStatus'
          shape={createRoundedBarShape('notAssessed')}
        >
          <LabelList
            dataKey='total'
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
