import { ChartConfig } from '@tumaet/prompt-ui-components'
import { SkillDistributionStats } from '../../../interfaces/surveyStatistics'
import {
  SkillLevel,
  SKILL_LEVEL_LABELS,
  SKILL_LEVEL_ORDER,
} from '../../../interfaces/skillResponse'
import { StackedBarChart } from './StackedBarChart'

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

const SKILL_LEVEL_COLORS: Record<SkillLevelValue, string> = {
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
    <StackedBarChart
      data={chartData}
      xKey='dataKey'
      seriesKeys={SKILL_LEVELS}
      seriesColors={SKILL_LEVEL_COLORS}
      chartConfig={chartConfig}
      totalKey='total'
      tooltip={<CustomTooltip />}
    />
  )
}
