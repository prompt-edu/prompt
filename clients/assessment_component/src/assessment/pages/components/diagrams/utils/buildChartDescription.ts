interface BucketPoint {
  dataKey: string
  count: number
}

interface GroupPoint {
  label: string
  average: number
  counts: Record<string, number>
}

const MAX_LISTED_GROUPS = 8

const students = (count: number): string => `${count} ${count === 1 ? 'student' : 'students'}`

const totalCount = (counts: Record<string, number>): number =>
  Object.values(counts).reduce((sum, n) => sum + n, 0)

export const describeBucketCounts = (points: BucketPoint[]): string => {
  if (points.length === 0) return 'No data available.'
  return `${points.map((p) => `${p.dataKey}: ${students(p.count)}`).join(', ')}.`
}

export const describeGroupAverages = (points: GroupPoint[]): string => {
  const withData = points.filter((p) => totalCount(p.counts) > 0)
  if (withData.length === 0) return 'No data available.'

  if (points.length <= MAX_LISTED_GROUPS) {
    return `${points
      .map((p) => {
        const total = totalCount(p.counts)
        return total === 0
          ? `${p.label}: no data`
          : `${p.label}: average ${p.average.toFixed(1)} (${students(total)})`
      })
      .join(', ')}.`
  }

  const averages = withData.map((p) => p.average)
  return `${points.length} groups. Individual averages range from ${Math.min(...averages).toFixed(
    1,
  )} to ${Math.max(...averages).toFixed(1)}.`
}
