import { useMemo } from 'react'
import { ChartConfig } from '@tumaet/prompt-ui-components'
import { PreferenceCount, TeamPopularityStats } from '../../../interfaces/surveyStatistics'
import { commonWordPrefix, ordinal } from '../utils/chartFormatters'
import { StackedBarChart } from './StackedBarChart'

interface TeamPopularityChartProps {
  data: TeamPopularityStats[]
}

type ChoiceKey = `choice${number}`

interface ChartRow extends TeamPopularityStats {
  displayName: string
  topChoiceTotal: number
  [key: ChoiceKey]: number
}

interface CustomTooltipProps {
  active?: boolean
  payload?: { payload: ChartRow }[]
}

const TOP_CHOICE_COLORS = ['#5eead4', '#7dd3fc', '#a5b4fc']
const MAX_TOP_CHOICES = TOP_CHOICE_COLORS.length

const CustomTooltip = ({ active, payload }: CustomTooltipProps) => {
  if (!active || !payload?.length) return null
  const { teamName, avgPreference, responseCount, preferenceCounts } = payload[0].payload
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
      {preferenceCounts.length > 0 && (
        <div className='mt-1.5 border-t pt-1.5 flex flex-col gap-0.5'>
          {preferenceCounts.map((pc) => (
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
  const { choiceKeys, seriesColors, chartConfig, chartData } = useMemo(() => {
    const numTeams = data.length
    const numTopChoices = Math.min(MAX_TOP_CHOICES, numTeams)
    const keys = Array.from({ length: numTopChoices }, (_, i) => `choice${i + 1}` as ChoiceKey)

    const colors = Object.fromEntries(keys.map((key, i) => [key, TOP_CHOICE_COLORS[i]])) as Record<
      ChoiceKey,
      string
    >

    const config: ChartConfig = Object.fromEntries(
      keys.map((key, i) => [key, { label: `${ordinal(i + 1)} choice`, color: colors[key] }]),
    )

    const sorted = [...data].sort((a, b) => {
      if (a.avgPreference === null && b.avgPreference === null) return 0
      if (a.avgPreference === null) return 1
      if (b.avgPreference === null) return -1
      return a.avgPreference - b.avgPreference
    })

    const namePrefix = commonWordPrefix(sorted.map((team) => team.teamName))

    const rows: ChartRow[] = sorted.map((team) => {
      const countsByRank = new Map((team.preferenceCounts ?? []).map((pc) => [pc.rank, pc.count]))
      const allCounts: PreferenceCount[] = Array.from({ length: numTeams }, (_, i) => ({
        rank: i + 1,
        count: countsByRank.get(i + 1) ?? 0,
      }))
      const row: ChartRow = {
        ...team,
        displayName: team.teamName.slice(namePrefix.length),
        preferenceCounts: allCounts,
        topChoiceTotal: 0,
      }
      keys.forEach((key, i) => {
        row[key] = countsByRank.get(i + 1) ?? 0
        row.topChoiceTotal += row[key]
      })
      return row
    })

    return { choiceKeys: keys, seriesColors: colors, chartConfig: config, chartData: rows }
  }, [data])

  return (
    <StackedBarChart
      data={chartData}
      xKey='displayName'
      seriesKeys={choiceKeys}
      seriesColors={seriesColors}
      chartConfig={chartConfig}
      totalKey='topChoiceTotal'
      tooltip={<CustomTooltip />}
    />
  )
}
