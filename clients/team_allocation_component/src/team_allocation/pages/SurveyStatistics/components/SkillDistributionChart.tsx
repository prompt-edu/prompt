import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartConfig } from '@tumaet/prompt-ui-components'
import { BarChart, Bar, LabelList, XAxis, YAxis, Rectangle, RectangleProps } from 'recharts'
import { SkillDistributionStats, SkillLevel } from '../../../interfaces/surveyStatistics'

interface SkillDistributionChartProps {
  data: SkillDistributionStats[]
}

const SKILL_LEVEL_ORDER: SkillLevel[] = ['very_bad', 'bad', 'ok', 'good', 'very_good']

// Exact same colors as the assessment component's chartConfig
const SKILL_LEVEL_COLORS: Record<SkillLevel, string> = {
  very_bad: '#fca5a5',  // red-300
  bad: '#fdba74',       // orange-300
  ok: '#fde68a',        // amber-200
  good: '#86efac',      // green-300
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

type CornerRadius = [number, number, number, number]

function getCornerRadius(payload: Record<string, number>, segmentKey: string): CornerRadius {
  const activeSegments = SKILL_LEVEL_ORDER.filter((k) => (payload[k] ?? 0) > 0)
  if (!activeSegments.includes(segmentKey as SkillLevel)) return [0, 0, 0, 0]
  const idx = activeSegments.indexOf(segmentKey as SkillLevel)
  const isTop = idx === activeSegments.length - 1
  const isBottom = idx === 0
  const r = 4
  return [isTop ? r : 0, isTop ? r : 0, isBottom ? r : 0, isBottom ? r : 0]
}

export const SkillDistributionChart = ({ data }: SkillDistributionChartProps) => {
  const chartData = data
    .map((skill) => ({
      dataKey: skill.skillName,
      ...Object.fromEntries(
        SKILL_LEVEL_ORDER.map((level) => [level, skill.levelCounts[level] ?? 0]),
      ),
      total: Object.values(skill.levelCounts).reduce((sum, v) => sum + v, 0),
    }))
    .sort((a, b) => a.dataKey.localeCompare(b.dataKey))

  const createRoundedShape = (level: SkillLevel) => {
    const Shape = (props: RectangleProps & { payload: Record<string, number> }) => {
      const { x, y, width, height, payload } = props
      const radius = getCornerRadius(payload, level)
      return (
        <Rectangle
          x={x}
          y={y}
          width={width}
          height={height}
          radius={radius}
          fill={SKILL_LEVEL_COLORS[level]}
        />
      )
    }
    Shape.displayName = `Shape(${level})`
    return Shape
  }

  return (
    <ChartContainer config={chartConfig} className='mx-auto w-full h-[280px]'>
      <BarChart data={chartData} margin={{ top: 30, right: 10, bottom: 0, left: 10 }}>
        <XAxis
          dataKey='dataKey'
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12 }}
          interval={0}
          height={50}
        />
        <YAxis hide />
        <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
        {SKILL_LEVEL_ORDER.map((level, index) => (
          <Bar key={level} dataKey={level} stackId='levels' fill={SKILL_LEVEL_COLORS[level]} shape={createRoundedShape(level)}>
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
