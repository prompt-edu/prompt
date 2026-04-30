import { ChartContainer, ChartTooltip, ChartConfig } from '@tumaet/prompt-ui-components'
import { BarChart, Bar, LabelList, XAxis, YAxis, Cell } from 'recharts'
import { PreferenceCount, TeamPopularityStats } from '../../../interfaces/surveyStatistics'

interface TeamPopularityChartProps {
  data: TeamPopularityStats[]
}

const getColor = (avg: number | null, min: number, max: number): string => {
  if (avg === null) return '#e2e8f0' // slate-200 — no responses
  if (max === min) return '#5eead4'
  const normalized = (avg - min) / (max - min)
  if (normalized < 0.2) return '#5eead4' // teal-300   — most popular
  if (normalized < 0.4) return '#7dd3fc' // sky-300
  if (normalized < 0.6) return '#a5b4fc' // indigo-300
  if (normalized < 0.8) return '#d8b4fe' // purple-300
  return '#f9a8d4' // pink-300   — least popular
}

const chartConfig: ChartConfig = {
  avgPreference: { label: 'Avg. choice rank' },
}

const ordinal = (n: number) => {
  const mod100 = n % 100
  if (mod100 >= 11 && mod100 <= 13) return `${n}th`
  switch (n % 10) {
    case 1:
      return `${n}st`
    case 2:
      return `${n}nd`
    case 3:
      return `${n}rd`
    default:
      return `${n}th`
  }
}

type TooltipRow = TeamPopularityStats & { fill: string; preferenceCounts: PreferenceCount[] }

interface CustomTooltipProps {
  active?: boolean
  payload?: { payload: TooltipRow }[]
}

const CustomTooltip = ({ active, payload }: CustomTooltipProps) => {
  if (!active || !payload?.length) return null
  const { teamName, avgPreference, responseCount, fill, preferenceCounts } = payload[0].payload
  const sorted: PreferenceCount[] = [...(preferenceCounts ?? [])].sort((a, b) => a.rank - b.rank)
  return (
    <div className='rounded-lg border bg-background px-3 py-2 text-sm shadow-md'>
      <p className='font-medium mb-1'>{teamName}</p>
      <div className='flex items-center gap-2'>
        <span className='h-2.5 w-2.5 shrink-0 rounded-[2px]' style={{ backgroundColor: fill }} />
        <span className='text-muted-foreground'>Avg. choice rank</span>
        <span className='ml-auto font-mono font-medium'>
          {avgPreference !== null ? avgPreference.toFixed(2) : '—'}
        </span>
      </div>
      <div className='flex items-center gap-2 mt-0.5'>
        <span className='h-2.5 w-2.5 shrink-0' />
        <span className='text-muted-foreground'>Responses</span>
        <span className='ml-auto font-mono font-medium'>{responseCount}</span>
      </div>
      {sorted.length > 0 && (
        <div className='mt-1.5 border-t pt-1.5 flex flex-col gap-0.5'>
          {sorted.map((pc) => (
            <div key={pc.rank} className='flex items-center gap-2'>
              <span className='h-2.5 w-2.5 shrink-0' />
              <span className='text-muted-foreground'>{ordinal(pc.rank)} choice</span>
              <span className='ml-auto font-mono font-medium'>{pc.count}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export const TeamPopularityChart = ({ data }: TeamPopularityChartProps) => {
  const numTeams = data.length
  // Sort rated teams first (ascending avg), unrated teams last
  const sorted = [...data].sort((a, b) => {
    if (a.avgPreference === null && b.avgPreference === null) return 0
    if (a.avgPreference === null) return 1
    if (b.avgPreference === null) return -1
    return a.avgPreference - b.avgPreference
  })
  const ratedTeams = sorted.filter((t) => t.avgPreference !== null)
  const min = ratedTeams[0]?.avgPreference ?? 1
  const max = ratedTeams[ratedTeams.length - 1]?.avgPreference ?? 1

  const chartData = sorted.map((team) => {
    const countsByRank = new Map((team.preferenceCounts ?? []).map((pc) => [pc.rank, pc.count]))
    const allCounts: PreferenceCount[] = Array.from({ length: numTeams }, (_, i) => ({
      rank: i + 1,
      count: countsByRank.get(i + 1) ?? 0,
    }))
    return { ...team, preferenceCounts: allCounts, fill: getColor(team.avgPreference, min, max) }
  })

  return (
    <ChartContainer config={chartConfig} className='mx-auto w-full h-[280px]'>
      <BarChart data={chartData} margin={{ top: 25, right: 10, bottom: 0, left: 10 }}>
        <XAxis
          dataKey='teamName'
          axisLine={false}
          tickLine={false}
          tick={{ fontSize: 12 }}
          interval={0}
          height={50}
        />
        <YAxis hide />
        <ChartTooltip cursor={false} content={<CustomTooltip />} />
        <Bar dataKey='avgPreference' radius={[4, 4, 0, 0]}>
          {chartData.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.fill} />
          ))}
          <LabelList
            dataKey='avgPreference'
            position='top'
            offset={8}
            className='fill-foreground'
            fontSize={12}
            formatter={(v) => (typeof v === 'number' ? v.toFixed(1) : '—')}
          />
        </Bar>
      </BarChart>
    </ChartContainer>
  )
}
