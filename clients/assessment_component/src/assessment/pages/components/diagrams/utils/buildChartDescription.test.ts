import assert from 'node:assert'

import { describeBucketCounts, describeGroupAverages } from './buildChartDescription'

assert.strictEqual(describeBucketCounts([]), 'No data available.')
assert.strictEqual(
  describeBucketCounts([
    { dataKey: '1.0', count: 4 },
    { dataKey: 'No Grade', count: 1 },
  ]),
  '1.0: 4 students, No Grade: 1 student.',
)

assert.strictEqual(
  describeGroupAverages([{ label: 'Team A', average: 0, counts: { ok: 0 } }]),
  'No data available.',
)
assert.strictEqual(
  describeGroupAverages([
    { label: 'Team A', average: 2.34, counts: { ok: 3 } },
    { label: 'Team B', average: 0, counts: { ok: 0 } },
  ]),
  'Team A: average 2.3 (3 students), Team B: no data.',
)

const many = Array.from({ length: 9 }, (_, i) => ({
  label: `G${i}`,
  average: 1 + i * 0.25,
  counts: { ok: 1 },
}))
assert.strictEqual(
  describeGroupAverages(many),
  '9 groups. Individual averages range from 1.0 to 3.0.',
)

console.log('buildChartDescription self-check passed')
