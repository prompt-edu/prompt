import {
  ChartContainer,
  ChartTooltip,
  ChartLegend,
  ChartLegendContent,
  ChartConfig,
  useIsMobile,
} from '@tumaet/prompt-ui-components'
import { BarChart, Bar, LabelList, XAxis, YAxis, Label, CartesianGrid } from 'recharts'
import { SkillDistributionStats } from '../../../interfaces/surveyStatistics'
import {
  SkillLevel,
  SKILL_LEVEL_LABELS,
  SKILL_LEVEL_ORDER,
} from '../../../interfaces/skillResponse'
import { createRoundedStackShape } from '../utils/roundedStackShape'
import { truncate } from '../utils/chartFormatters'

interface SkillDistributionChartProps {
  data: SkillDistributionStats[]
}

type SkillLevelValue = `${SkillLevel}`

interface SkillChartRow extends Record<SkillLevelValue, number> {
  dataKey: string
  total: number
}

interface CustomTooltipProps {
  active?: boolean
  payload?: { payload: SkillChartRow }[]
}

const SKILL_LEVELS: SkillLevelValue[] = SKILL_LEVEL_ORDER

const SKILL_LEVEL_COLORS: Record<SkillLevel, string> = {
  [SkillLevel.VERY_BAD]: '#fca5a5',
  [SkillLevel.BAD]: '#fdba74',
  [SkillLevel.OK]: '#fde68a',
  [SkillLevel.GOOD]: '#86efac',
  [SkillLevel.VERY_GOOD]: '#93c5fd',
}

const chartConfig: ChartConfig = Object.fromEntries(
  SKILL_LEVELS.map((level) => [
    level,
    { label: SKILL_LEVEL_LABELS[level], color: SKILL_LEVEL_COLORS[level] },
  ]),
)

const SKILL_LEVEL_SHAPES = Object.fromEntries(
  SKILL_LEVELS.map((level) => [
    level,
    createRoundedStackShape(SKILL_LEVELS, level, SKILL_LEVEL_COLORS[level]),
  ]),
)

const CustomTooltip = ({ active, payload }: CustomTooltipProps) => {
  if (!active || !payload?.length) return null
  const row = payload[0].payload
  return (
    <div className='rounded-lg border bg-background px-3 py-2 text-sm shadow-md'>
      <p className='font-medium mb-1'>{row.dataKey}</p>
      <div className='flex flex-col gap-0.5'>
        {SKILL_LEVELS.map((level) => {
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
  const isMobile = useIsMobile()

  const chartData: SkillChartRow[] = data.map((skill) => {
    const levelCounts = Object.fromEntries(
      SKILL_LEVELS.map((level) => [level, skill.levelCounts[level] ?? 0]),
    ) as Record<SkillLevel, number>
    return {
      ...levelCounts,
      dataKey: skill.skillName,
      total: SKILL_LEVELS.reduce((sum, level) => sum + levelCounts[level], 0),
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
        <ChartTooltip cursor={false} content={<CustomTooltip />} />
        <ChartLegend
          content={<ChartLegendContent />}
          itemSorter={(item) => SKILL_LEVELS.indexOf(String(item.dataKey) as SkillLevelValue)}
        />
        {SKILL_LEVELS.map((level, index) => (
          <Bar
            key={level}
            dataKey={level}
            stackId='levels'
            fill={SKILL_LEVEL_COLORS[level]}
            shape={SKILL_LEVEL_SHAPES[level]}
          >
            {index === SKILL_LEVELS.length - 1 && (
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
