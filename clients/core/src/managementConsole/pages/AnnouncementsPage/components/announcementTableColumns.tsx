import {
  type AnnouncementStatus,
  formatAnnouncementDate,
  getAnnouncementStatus,
} from '@core/announcementBanner/utils/announcementStatus'
import type { Announcement } from '@core/interfaces/announcement'
import { Badge, type PromptTableColumnDef, Switch } from '@tumaet/prompt-ui-components'
import { SEVERITY_LABELS, STATUS_LABELS } from './announcementLabels'

const STATUS_BADGE_VARIANTS: Record<
  AnnouncementStatus,
  'default' | 'secondary' | 'outline' | 'destructive'
> = {
  active: 'default',
  scheduled: 'secondary',
  disabled: 'outline',
  expired: 'outline',
}

export const getAnnouncementTableColumns = (
  onToggleEnabled: (announcement: Announcement, enabled: boolean) => void,
  now: Date,
): PromptTableColumnDef<Announcement>[] => [
  {
    id: 'status',
    header: 'Status',
    accessorFn: (announcement) => STATUS_LABELS[getAnnouncementStatus(announcement, now)],
    cell: ({ row }) => {
      const status = getAnnouncementStatus(row.original, now)
      return <Badge variant={STATUS_BADGE_VARIANTS[status]}>{STATUS_LABELS[status]}</Badge>
    },
  },
  {
    id: 'severity',
    header: 'Severity',
    accessorFn: (announcement) => SEVERITY_LABELS[announcement.severity],
  },
  {
    id: 'announcement',
    header: 'Announcement',
    accessorFn: (announcement) => `${announcement.title} ${announcement.message}`,
    cell: ({ row }) => (
      <div className='flex max-w-72 flex-col 2xl:max-w-md'>
        {row.original.title && <span className='font-medium'>{row.original.title}</span>}
        <span className='truncate text-muted-foreground'>{row.original.message}</span>
      </div>
    ),
  },
  {
    id: 'startsAt',
    header: 'Starts',
    accessorFn: (announcement) => announcement.startsAt ?? '',
    cell: ({ row }) =>
      row.original.startsAt ? formatAnnouncementDate(row.original.startsAt) : 'Immediately',
  },
  {
    id: 'expiresAt',
    header: 'Expires',
    accessorFn: (announcement) => announcement.expiresAt ?? '',
    cell: ({ row }) =>
      row.original.expiresAt ? formatAnnouncementDate(row.original.expiresAt) : 'Never',
  },
  {
    id: 'enabled',
    header: 'Enabled',
    accessorFn: (announcement) => announcement.enabled,
    cell: ({ row }) => (
      <Switch
        checked={row.original.enabled}
        onClick={(event) => event.stopPropagation()}
        onCheckedChange={(enabled) => onToggleEnabled(row.original, enabled)}
        aria-label={row.original.enabled ? 'Disable announcement' : 'Enable announcement'}
      />
    ),
  },
]
