import { ReactElement, useMemo } from 'react'
import {
  ChartContainer,
  ChartTooltip,
  ChartLegend,
  ChartLegendContent,
  ChartConfig,
  useIsMobile,
} from '@tumaet/prompt-ui-components'
import { BarChart, Bar, LabelList, XAxis, YAxis, Label, CartesianGrid } from 'recharts'
import { createRoundedStackShape } from '../utils/roundedStackShape'
import { truncate } from '../utils/chartFormatters'

interface StackedBarChartProps<Row extends object, SeriesKey extends string> {
  data: Row[]
  xKey: string
  seriesKeys: readonly SeriesKey[]
  seriesColors: Record<SeriesKey, string>
  chartConfig: ChartConfig
  totalKey: string
  tooltip: ReactElement
}

export const StackedBarChart = <Row extends object, SeriesKey extends string>({
  data,
  xKey,
  seriesKeys,
  seriesColors,
  chartConfig,
  totalKey,
  tooltip,
}: StackedBarChartProps<Row, SeriesKey>) => {
  const isMobile = useIsMobile()

  const shapes = useMemo(
    () =>
      Object.fromEntries(
        seriesKeys.map((key) => [key, createRoundedStackShape(seriesKeys, key, seriesColors[key])]),
      ),
    [seriesKeys, seriesColors],
  )

  return (
    <ChartContainer config={chartConfig} className='mx-auto w-full h-[320px]'>
      <BarChart data={data} margin={{ top: 30, right: 10, bottom: 0, left: 10 }}>
        <CartesianGrid
          horizontal={true}
          vertical={false}
          strokeDasharray='5 5'
          stroke='#e5e7eb'
          opacity={1}
        />
        <XAxis
          dataKey={xKey}
          axisLine={false}
          tickLine={false}
          tick={isMobile ? { fontSize: 10, angle: -30, textAnchor: 'end' } : { fontSize: 12 }}
          tickFormatter={(value: string) => truncate(value, isMobile ? 10 : 12)}
          interval={0}
          height={isMobile ? 60 : 50}
        />
        <YAxis
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12, fill: '#a3a3a3' }}
          allowDecimals={false}
          width={isMobile ? 30 : 60}
        >
          {!isMobile && <Label value='Students' angle={-90} position='insideLeft' fill='#a3a3a3' />}
        </YAxis>
        <ChartTooltip cursor={false} content={tooltip} />
        <ChartLegend
          content={<ChartLegendContent />}
          itemSorter={(item) => seriesKeys.indexOf(String(item.dataKey) as SeriesKey)}
        />
        {seriesKeys.map((key, index) => (
          <Bar key={key} dataKey={key} stackId='stack' fill={seriesColors[key]} shape={shapes[key]}>
            {index === seriesKeys.length - 1 && (
              <LabelList
                dataKey={totalKey}
                position='top'
                offset={10}
                className='fill-foreground'
                fontSize={12}
              />
            )}
          </Bar>
        ))}
      </BarChart>
    </ChartContainer>
  )
}
