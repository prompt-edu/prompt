import { type BarShapeProps, Rectangle } from 'recharts'

type CornerRadius = [number, number, number, number]

const RADIUS = 4

const getCornerRadius = (
  payload: Record<string, number> | undefined,
  stackOrder: readonly string[],
  segmentKey: string,
): CornerRadius => {
  if (!payload) return [0, 0, 0, 0]
  const activeSegments = stackOrder.filter((key) => (payload[key] ?? 0) > 0)
  const idx = activeSegments.indexOf(segmentKey)
  if (idx === -1) return [0, 0, 0, 0]
  const isTop = idx === activeSegments.length - 1
  const isBottom = idx === 0
  return [isTop ? RADIUS : 0, isTop ? RADIUS : 0, isBottom ? RADIUS : 0, isBottom ? RADIUS : 0]
}

export const createRoundedStackShape = (
  stackOrder: readonly string[],
  segmentKey: string,
  fill: string,
) => {
  const Shape = (props: BarShapeProps) => {
    const { x, y, width, height, payload } = props
    const radius = getCornerRadius(
      payload as Record<string, number> | undefined,
      stackOrder,
      segmentKey,
    )
    return <Rectangle x={x} y={y} width={width} height={height} radius={radius} fill={fill} />
  }
  Shape.displayName = `RoundedStackShape(${segmentKey})`
  return Shape
}
