import type { DataPoint } from '../interfaces/DataPoint'

type CornerRadius = [number, number, number, number]

// Default keys for stacked charts
const DEFAULT_STACK_KEYS = ['accepted', 'rejected', 'notAssessed']

/**
 * Automatically detects whether we're in a stacked chart based on presence in stackKeys.
 * Returns rounded corners accordingly.
 */
export function getCornerRadius(
  dataPoint: DataPoint,
  segmentKey: string,
  stackKeys: string[] = DEFAULT_STACK_KEYS,
): CornerRadius {
  const radius = 4
  const isStacked = stackKeys.includes(segmentKey)

  if (!isStacked) {
    // Non-stacked: apply top-corner rounding to all bars
    return [radius, radius, 0, 0]
  }

  if (!dataPoint) {
    return [0, 0, 0, 0]
  }

  const activeSegments = stackKeys.filter((key) => (dataPoint[key] ?? 0) > 0)

  if (!activeSegments.includes(segmentKey)) {
    return [0, 0, 0, 0]
  }

  const segmentIndex = activeSegments.indexOf(segmentKey)
  const isOnlySegment = activeSegments.length === 1
  const isTopSegment = segmentIndex === activeSegments.length - 1
  const isBottomSegment = segmentIndex === 0

  return [
    isTopSegment || isOnlySegment ? radius : 0,
    isTopSegment || isOnlySegment ? radius : 0,
    isBottomSegment || isOnlySegment ? radius : 0,
    isBottomSegment || isOnlySegment ? radius : 0,
  ]
}
