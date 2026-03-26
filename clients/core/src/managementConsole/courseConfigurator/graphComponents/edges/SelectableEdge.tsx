import { BaseEdge, getBezierPath, type EdgeProps } from '@xyflow/react'

export function SelectableEdge({
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
  selected,
}: EdgeProps) {
  const [edgePath] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  })

  return (
    <>
      {selected && (
        <path
          d={edgePath}
          fill='none'
          stroke={(style.stroke as string) || '#000'}
          strokeWidth={((style.strokeWidth as number) || 2) + 5}
          strokeOpacity={0.3}
          strokeLinecap='round'
        />
      )}
      <BaseEdge path={edgePath} markerEnd={markerEnd} style={style} />
    </>
  )
}
