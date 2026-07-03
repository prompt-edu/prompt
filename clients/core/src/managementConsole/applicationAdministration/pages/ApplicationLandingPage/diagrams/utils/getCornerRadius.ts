import type { DataPoint } from '../../interfaces/DataPoint'

type CornerRadius = [number, number, number, number]

// Order the keys from bottom (first rendered) to top (last rendered)
const STACK_KEYS = ['accepted', 'rejected', 'notAssessed']

/**
 * Calculates the corner radius for a given segment in a stacked bar.
 * - If the segment is the only one in the bar, all corners are rounded.
 * - If multiple segments exist, only the top segment gets its top corners rounded,
 *   and only the bottom segment gets its bottom corners rounded.
 *
 * @param dataPoint The data point for the current bar.
 * @param segmentKey The key for the current segment (e.g. 'accepted').
 * @param stackKeys The ordered keys in the stack (default is STACK_KEYS).
 * @returns An array of four numbers representing the radius for the corners in the order:
 * [top-left, top-right, bottom-right, bottom-left].
 */
export function getCornerRadius(dataPoint: DataPoint, segmentKey: string): CornerRadius {
  // Only consider segments that have a positive (non-zero) value.
  const activeSegments = STACK_KEYS.filter((key) => (dataPoint[key] ?? 0) > 0)

  // If the segment isn't present, no rounding.
  if (!activeSegments.includes(segmentKey)) {
    return [0, 0, 0, 0]
  }

  const segmentIndex = activeSegments.indexOf(segmentKey)
  const isOnlySegment = activeSegments.length === 1
  const isTopSegment = segmentIndex === activeSegments.length - 1
  const isBottomSegment = segmentIndex === 0
  const radius = 4

  return [
    // top-left corner: round if this segment is at the top or the only segment
    isTopSegment || isOnlySegment ? radius : 0,
    // top-right corner
    isTopSegment || isOnlySegment ? radius : 0,
    // bottom-right corner: round if this segment is at the bottom or the only segment
    isBottomSegment || isOnlySegment ? radius : 0,
    // bottom-left corner
    isBottomSegment || isOnlySegment ? radius : 0,
  ]
}
