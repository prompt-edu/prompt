import type { ChartConfig } from '@tumaet/prompt-ui-components'

export const chartConfig: ChartConfig = {
  notAssessed: {
    label: 'Not Assessed',
    color: 'hsl(var(--muted))',
  },
  accepted: {
    label: 'Accepted',
    color: 'hsl(var(--success))',
  },
  rejected: {
    label: 'Rejected',
    color: 'hsl(var(--destructive))',
  },
}
