import type { Announcement } from '@core/interfaces/announcement'
import type { RowAction } from '@tumaet/prompt-ui-components'
import { Pencil, Trash2 } from 'lucide-react'

interface AnnouncementTableActionsProps {
  onEdit: (announcement: Announcement) => void
  onDelete: (announcements: Announcement[]) => void
}

export const getAnnouncementTableActions = ({
  onEdit,
  onDelete,
}: AnnouncementTableActionsProps): RowAction<Announcement>[] => [
  {
    label: 'Edit',
    icon: <Pencil />,
    onAction: (rows) => onEdit(rows[0]),
    hide: (rows) => rows.length > 1,
  },
  {
    label: 'Delete',
    icon: <Trash2 />,
    onAction: (rows) => onDelete(rows),
    confirm: {
      title: 'Delete Announcement',
      description: (count) =>
        `Are you sure you want to permanently delete ${count} announcement${count > 1 ? 's' : ''}?`,
      confirmLabel: 'Delete',
      variant: 'destructive',
    },
  },
]
