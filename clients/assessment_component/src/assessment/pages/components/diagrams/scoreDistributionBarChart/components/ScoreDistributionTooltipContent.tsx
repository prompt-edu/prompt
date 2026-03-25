import { getLevelConfig } from '@tumaet/prompt-ui-components'
import { ScoreLevel } from '@tumaet/prompt-shared-state'

export function ScoreDistributionTooltipContent(props: any) {
  if (!props.active || !props.payload || !props.payload[0]) {
    return undefined
  }

  const data = props.payload[0].payload
  const counts: Record<ScoreLevel, number> = data.counts
  const total = Object.values(counts).reduce((sum, count) => sum + count, 0)

  if (total === 0) {
    return (
      <div className='rounded-lg border bg-background p-2 shadow-md'>
        <div className='font-medium mb-2'>{data.name}</div>
        <div className='text-muted-foreground'>No data available</div>
      </div>
    )
  }

  return (
    <div className='rounded-lg border bg-background p-2 shadow-md'>
      <div className='font-medium mb-2'>{data.name}</div>
      <div className='space-y-1 text-sm'>
        <div className='grid grid-cols-2 gap-x-3'>
          <span className='text-muted-foreground'>Average:</span>
          <span className='font-medium'>{data.average.toFixed(1)}</span>
        </div>
        <div className='grid grid-cols-2 gap-x-3'>
          <span className='text-muted-foreground'>Median:</span>
          <span className='font-medium'>{getLevelConfig(data.median as ScoreLevel).title}</span>
        </div>
        <div className='grid grid-cols-2 gap-x-3'>
          <span className='text-muted-foreground'>Lower Quartile:</span>
          <span className='font-medium'>{data.lowerQuartile.toFixed(1)}</span>
        </div>
        <div className='grid grid-cols-2 gap-x-3'>
          <span className='text-muted-foreground'>Upper Quartile:</span>
          <span className='font-medium'>{data.upperQuartile.toFixed(1)}</span>
        </div>
        <div className='h-px bg-border my-2'></div>
        <div className='font-medium'>Distribution</div>
        {Object.entries(counts).map(([key, value]) => (
          <div key={key} className='grid grid-cols-2 gap-x-3'>
            <span className='text-muted-foreground'>
              {getLevelConfig(key as ScoreLevel).title}:
            </span>
            <span className='font-medium'>
              {value} ({((value / total) * 100).toFixed(0)}%)
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
