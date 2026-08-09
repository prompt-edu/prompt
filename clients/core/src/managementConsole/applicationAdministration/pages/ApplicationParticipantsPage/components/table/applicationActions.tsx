import { PassStatus } from '@tumaet/prompt-shared-state'
import type { RowAction } from '@tumaet/prompt-ui-components'
import { CheckCircle, Download, FileUser, MailCheck, MailX, Trash2, XCircle } from 'lucide-react'
import type { ApplicationRow } from './applicationRow'

const plural = (count: number) => (count === 1 ? '' : 's')

const allHaveStatus = (rows: ApplicationRow[], status: PassStatus) =>
  rows.length > 0 && rows.every((row) => row.passStatus === status)

export function getApplicationActions(
  deleteApplications: (ids: string[]) => void,
  onView: (row: ApplicationRow) => void,
  actions: {
    setPassed: (rows: ApplicationRow[]) => void
    setFailed: (rows: ApplicationRow[]) => void
    sendAcceptanceMail: (rows: ApplicationRow[]) => void
    sendRejectionMail: (rows: ApplicationRow[]) => void
    exportCsv: (rows: ApplicationRow[]) => void | Promise<void>
  },
): RowAction<ApplicationRow>[] {
  return [
    {
      label: 'Export CSV',
      icon: <Download className='h-4 w-4' />,
      onAction: actions.exportCsv,
    },
    {
      label: 'Accept',
      icon: <CheckCircle className='h-4 w-4' />,
      confirm: {
        title: 'Confirm',
        description: (c) => `Accept ${c} applicant${plural(c)} to the course?`,
        confirmLabel: 'Accept',
      },
      onAction: actions.setPassed,
    },
    {
      label: 'Reject',
      icon: <XCircle className='h-4 w-4' />,
      confirm: {
        title: 'Confirm',
        description: (c) => `Reject ${c} applicant${plural(c)}?`,
        confirmLabel: 'Reject',
        variant: 'destructive',
      },
      onAction: actions.setFailed,
    },
    {
      label: 'Send Acceptance Mail',
      icon: <MailCheck className='h-4 w-4' />,
      confirm: {
        title: 'Send acceptance mail',
        description: (c) => `Send the acceptance mail to ${c} accepted applicant${plural(c)}?`,
        confirmLabel: 'Send',
      },
      disabled: (rows) => !allHaveStatus(rows, PassStatus.PASSED),
      onAction: actions.sendAcceptanceMail,
    },
    {
      label: 'Send Rejection Mail',
      icon: <MailX className='h-4 w-4' />,
      confirm: {
        title: 'Send rejection mail',
        description: (c) => `Send the rejection mail to ${c} rejected applicant${plural(c)}?`,
        confirmLabel: 'Send',
        variant: 'destructive',
      },
      disabled: (rows) => !allHaveStatus(rows, PassStatus.FAILED),
      onAction: actions.sendRejectionMail,
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
          `Are you sure you want to delete ${count} application${plural(count)}?`,
        confirmLabel: 'Delete',
        variant: 'destructive',
      },
    },
  ]
}
