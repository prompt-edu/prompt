import { RowAction } from '@tumaet/prompt-ui-components'
import { ApplicationRow } from './applicationRow'
import { CheckCircle, FileUser, Trash2, XCircle } from 'lucide-react'

export function getApplicationActions(
  deleteApplications: (ids: string[]) => void,
  onView: (row: ApplicationRow) => void,
  actions: {
    setPassed: (rows: ApplicationRow[]) => void
    setFailed: (rows: ApplicationRow[]) => void
  },
): RowAction<ApplicationRow>[] {
  return [
    {
      label: 'Accept',
      icon: <CheckCircle className='h-4 w-4' />,
      confirm: {
        title: 'Confirm',
        description: (c) => `Accept ${c} applicant${c > 1 ? 's' : ''} to the course?`,
        confirmLabel: 'Accept',
      },
      onAction: actions.setPassed,
    },
    {
      label: 'Reject',
      icon: <XCircle className='h-4 w-4' />,
      confirm: {
        title: 'Confirm',
        description: (c) => `Reject ${c} applicant${c > 1 ? 's' : ''}?`,
        confirmLabel: 'Reject',
        variant: 'destructive',
      },
      onAction: actions.setFailed,
    },
    {
      label: 'View Application',
      icon: <FileUser className='h-4 w-4' />,
      onAction: ([row]) => onView(row),
      hide: (rows) => rows.length !== 1,
    },
    {
      label: 'Delete Application',
      icon: <Trash2 className='h-4 w-4 text-red-600' />,
      onAction: async (rows) => deleteApplications(rows.map((r) => r.courseParticipationID)),
      confirm: {
        title: 'Confirm Deletion',
        description: (count) =>
          `Are you sure you want to delete ${count} application${count > 1 ? 's' : ''}?`,
        confirmLabel: 'Delete',
        variant: 'destructive',
      },
    },
  ]
}
