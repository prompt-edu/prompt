import {
  ChartContainer,
  ChartTooltip,
  ChartLegend,
  ChartLegendContent,
  ChartConfig,
} from '@tumaet/prompt-ui-components'
import { BarChart, Bar, LabelList, XAxis, YAxis, Label, CartesianGrid } from 'recharts'
import { SkillDistributionStats, SkillLevel } from '../../../interfaces/surveyStatistics'
import { createRoundedStackShape } from '../utils/roundedStackShape'
import { truncate } from '../utils/chartFormatters'

interface SkillDistributionChartProps {
  data: SkillDistributionStats[]
}

interface SkillChartRow extends Record<SkillLevel, number> {
  dataKey: string
  total: number
}

interface CustomTooltipProps {
  active?: boolean
  payload?: { payload: SkillChartRow }[]
}

const SKILL_LEVEL_ORDER: SkillLevel[] = ['very_bad', 'bad', 'ok', 'good', 'very_good']

// Exact same colors as the assessment component's chartConfig
const SKILL_LEVEL_COLORS: Record<SkillLevel, string> = {
  very_bad: '#fca5a5', // red-300
  bad: '#fdba74', // orange-300
  ok: '#fde68a', // amber-200
  good: '#86efac', // green-300
  very_good: '#93c5fd', // blue-300
}

const SKILL_LEVEL_LABELS: Record<SkillLevel, string> = {
  very_bad: 'Very Bad',
  bad: 'Bad',
  ok: 'Ok',
  good: 'Good',
  very_good: 'Very Good',
}

const chartConfig: ChartConfig = Object.fromEntries(
  SKILL_LEVEL_ORDER.map((level) => [
    level,
    { label: SKILL_LEVEL_LABELS[level], color: SKILL_LEVEL_COLORS[level] },
  ]),
)

const CustomTooltip = ({ active, payload }: CustomTooltipProps) => {
  if (!active || !payload?.length) return null
  const row = payload[0].payload
  return (
    <div className='rounded-lg border bg-background px-3 py-2 text-sm shadow-md'>
      <p className='font-medium mb-1'>{row.dataKey}</p>
      <div className='flex flex-col gap-0.5'>
        {SKILL_LEVEL_ORDER.map((level) => {
          const count = row[level]
          const percentage = row.total > 0 ? Math.round((count / row.total) * 100) : 0
          return (
            <div key={level} className='flex items-center gap-2'>
              <span
                className='h-2.5 w-2.5 shrink-0 rounded-[2px]'
                style={{ backgroundColor: SKILL_LEVEL_COLORS[level] }}
              />
              <span className='text-muted-foreground'>{SKILL_LEVEL_LABELS[level]}</span>
              <span className='ml-auto pl-4 font-mono font-medium'>
                {count} <span className='text-muted-foreground'>({percentage}%)</span>
              </span>
            </div>
          )
        })}
      </div>
      <div className='mt-1.5 border-t pt-1.5 flex items-center gap-2'>
        <span className='h-2.5 w-2.5 shrink-0' />
        <span className='text-muted-foreground'>Responses</span>
        <span className='ml-auto pl-4 font-mono font-medium'>{row.total}</span>
      </div>
    </div>
  )
}

export const SkillDistributionChart = ({ data }: SkillDistributionChartProps) => {
  const chartData: SkillChartRow[] = data.map((skill) => {
    const levelCounts = Object.fromEntries(
      SKILL_LEVEL_ORDER.map((level) => [level, skill.levelCounts[level] ?? 0]),
    ) as Record<SkillLevel, number>
    return {
      ...levelCounts,
      dataKey: skill.skillName,
      total: SKILL_LEVEL_ORDER.reduce((sum, level) => sum + levelCounts[level], 0),
    }
  })

  return (
    <ChartContainer config={chartConfig} className='mx-auto w-full h-[320px]'>
      <BarChart data={chartData} margin={{ top: 30, right: 10, bottom: 0, left: 10 }}>
        <CartesianGrid
          horizontal={true}
          vertical={false}
          strokeDasharray='5 5'
          stroke='#e5e7eb'
          opacity={1}
        />
        <XAxis
          dataKey='dataKey'
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12 }}
          tickFormatter={truncate}
          interval={0}
          height={50}
        />
        <YAxis
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12, fill: '#a3a3a3' }}
          allowDecimals={false}
        >
          <Label value='Students' angle={-90} position='insideLeft' fill='#a3a3a3' />
        </YAxis>
        <ChartTooltip cursor={false} content={<CustomTooltip />} />
        <ChartLegend content={<ChartLegendContent />} />
        {SKILL_LEVEL_ORDER.map((level, index) => (
          <Bar
            key={level}
            dataKey={level}
            stackId='levels'
            fill={SKILL_LEVEL_COLORS[level]}
            shape={createRoundedStackShape(SKILL_LEVEL_ORDER, level, SKILL_LEVEL_COLORS[level])}
          >
            {index === SKILL_LEVEL_ORDER.length - 1 && (
              <LabelList
                dataKey='total'
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
