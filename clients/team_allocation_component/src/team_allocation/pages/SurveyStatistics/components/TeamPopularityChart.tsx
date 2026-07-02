import {
  ChartContainer,
  ChartTooltip,
  ChartLegend,
  ChartLegendContent,
  ChartConfig,
} from '@tumaet/prompt-ui-components'
import { BarChart, Bar, LabelList, XAxis, YAxis, Label, CartesianGrid } from 'recharts'
import { PreferenceCount, TeamPopularityStats } from '../../../interfaces/surveyStatistics'
import { createRoundedStackShape } from '../utils/roundedStackShape'
import { ordinal, truncate } from '../utils/chartFormatters'

interface TeamPopularityChartProps {
  data: TeamPopularityStats[]
}

interface ChartRow extends TeamPopularityStats {
  topChoiceTotal: number
  [key: `choice${number}`]: number
}

interface CustomTooltipProps {
  active?: boolean
  payload?: { payload: ChartRow }[]
}

const TOP_CHOICE_COLORS = ['#5eead4', '#7dd3fc', '#a5b4fc'] // teal-300, sky-300, indigo-300
const MAX_TOP_CHOICES = TOP_CHOICE_COLORS.length

const CustomTooltip = ({ active, payload }: CustomTooltipProps) => {
  if (!active || !payload?.length) return null
  const { teamName, avgPreference, responseCount, preferenceCounts } = payload[0].payload
  const sorted: PreferenceCount[] = [...(preferenceCounts ?? [])].sort((a, b) => a.rank - b.rank)
  return (
    <div className='rounded-lg border bg-background px-3 py-2 text-sm shadow-md'>
      <p className='font-medium mb-1'>{teamName}</p>
      <div className='flex items-center gap-2'>
        <span className='h-2.5 w-2.5 shrink-0' />
        <span className='text-muted-foreground'>Avg. choice rank</span>
        <span className='ml-auto pl-4 font-mono font-medium'>
          {avgPreference !== null ? avgPreference.toFixed(2) : '—'}
        </span>
      </div>
      <div className='flex items-center gap-2 mt-0.5'>
        <span className='h-2.5 w-2.5 shrink-0' />
        <span className='text-muted-foreground'>Responses</span>
        <span className='ml-auto pl-4 font-mono font-medium'>{responseCount}</span>
      </div>
      {sorted.length > 0 && (
        <div className='mt-1.5 border-t pt-1.5 flex flex-col gap-0.5'>
          {sorted.map((pc) => (
            <div key={pc.rank} className='flex items-center gap-2'>
              <span
                className='h-2.5 w-2.5 shrink-0 rounded-[2px]'
                style={{
                  backgroundColor:
                    pc.rank <= MAX_TOP_CHOICES ? TOP_CHOICE_COLORS[pc.rank - 1] : 'transparent',
                }}
              />
              <span className='text-muted-foreground'>{ordinal(pc.rank)} choice</span>
              <span className='ml-auto pl-4 font-mono font-medium'>{pc.count}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export const TeamPopularityChart = ({ data }: TeamPopularityChartProps) => {
  const numTeams = data.length
  const numTopChoices = Math.min(MAX_TOP_CHOICES, numTeams)
  const choiceKeys = Array.from({ length: numTopChoices }, (_, i) => `choice${i + 1}` as const)

  const chartConfig: ChartConfig = Object.fromEntries(
    choiceKeys.map((key, i) => [
      key,
      { label: `${ordinal(i + 1)} choice`, color: TOP_CHOICE_COLORS[i] },
    ]),
  )

  const sorted = [...data].sort((a, b) => {
    if (a.avgPreference === null && b.avgPreference === null) return 0
    if (a.avgPreference === null) return 1
    if (b.avgPreference === null) return -1
    return a.avgPreference - b.avgPreference
  })

  const chartData: ChartRow[] = sorted.map((team) => {
    const countsByRank = new Map((team.preferenceCounts ?? []).map((pc) => [pc.rank, pc.count]))
    const allCounts: PreferenceCount[] = Array.from({ length: numTeams }, (_, i) => ({
      rank: i + 1,
      count: countsByRank.get(i + 1) ?? 0,
    }))
    const row: ChartRow = {
      ...team,
      preferenceCounts: allCounts,
      topChoiceTotal: 0,
    }
    choiceKeys.forEach((key, i) => {
      row[key] = countsByRank.get(i + 1) ?? 0
      row.topChoiceTotal += row[key]
    })
    return row
  })

  return (
    <ChartContainer config={chartConfig} className='mx-auto w-full h-[320px]'>
      <BarChart data={chartData} margin={{ top: 25, right: 10, bottom: 0, left: 10 }}>
        <CartesianGrid
          horizontal={true}
          vertical={false}
          strokeDasharray='5 5'
          stroke='#e5e7eb'
          opacity={1}
        />
        <XAxis
          dataKey='teamName'
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
        {choiceKeys.map((key, index) => (
          <Bar
            key={key}
            dataKey={key}
            stackId='choices'
            fill={TOP_CHOICE_COLORS[index]}
            shape={createRoundedStackShape(choiceKeys, key, TOP_CHOICE_COLORS[index])}
          >
            {index === choiceKeys.length - 1 && (
              <LabelList
                dataKey='topChoiceTotal'
                position='top'
                offset={8}
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
