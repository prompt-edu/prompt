import { PassStatus } from '@tumaet/prompt-shared-state'
import { TableFilter } from '@tumaet/prompt-ui-components'
import { getApplicationStatusBadge, getApplicationStatusString } from './getApplicationStatusBadge'

export function getApplicationFilters(
  additionalScores?: { key: string; name: string }[],
): TableFilter[] {
  return [
    {
      type: 'select',
      id: 'passStatus',
      label: 'Status',
      options: Object.values(PassStatus),
      optionLabel: (v) => getApplicationStatusBadge(v as PassStatus),
      badge: {
        label: 'Application',
        displayValue(filtervalue) {
          const ps = filtervalue as PassStatus
          return getApplicationStatusString(ps)
        },
      },
    },
    {
      type: 'select',
      id: 'gender',
      label: 'Gender',
      options: ['male', 'female', 'diverse'],
    },
    {
      type: 'numericRange',
      id: 'score',
      label: 'Score',
      noValueLabel: 'No score',
    },
    ...(additionalScores ?? []).map((s) => ({
      type: 'numericRange' as const,
      id: s.key,
      label: s.name,
    })),
  ]
}
