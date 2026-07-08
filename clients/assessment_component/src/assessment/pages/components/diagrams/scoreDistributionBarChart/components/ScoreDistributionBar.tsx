import { mapNumberToScoreLevel, ScoreLevel } from '@tumaet/prompt-shared-state'
import { Rectangle } from 'recharts'

import { getBarColor } from '../../utils/chartConfig'

export const ScoreDistributionBar = (props: any) => {
  const { x = 0, y = 0, width = 0, height = 0, payload } = props
  if (payload.lowerQuartile === 0 || payload.upperQuartile === 0) return undefined

  const scale = height / (Object.keys(ScoreLevel).length - 1)
  const barColor = getBarColor(mapNumberToScoreLevel(payload.average))
  const upperQuartileY = y + height - (payload.upperQuartile - 1) * scale

  const minHeight = 6
  if (payload.upperQuartile - payload.lowerQuartile <= 0) {
    return (
      <Rectangle
        x={x}
        y={upperQuartileY - minHeight / 2}
        width={width}
        height={minHeight}
        fill={barColor}
        radius={[4, 4, 4, 4]}
      />
    )
  }

  const lowerQuartileY = y + height - (payload.lowerQuartile - 1) * scale
  const averageY = y + height - (payload.average - 1) * scale
  const rectHeight = lowerQuartileY - upperQuartileY

  return (
    <g>
      <Rectangle
        x={x}
        y={upperQuartileY}
        width={width}
        height={rectHeight}
        fill={barColor}
        radius={[4, 4, 4, 4]}
      />
      <line
        x1={x}
        y1={averageY}
        x2={x + width}
        y2={averageY}
        stroke='#ffffff'
        strokeWidth={2}
        strokeOpacity={0.8}
      />
    </g>
  )
}
