import { ExtraParticipantColumn, ParticipantRow } from './participationRow'
import { CheckCircle, Download, XCircle } from 'lucide-react'
import { downloadParticipations } from '../utils/downloadParticipations'
import { RowAction } from '@tumaet/prompt-ui-components'

export interface ExportDeps {
  prevDataKeys?: string[]
  restrictedDataKeys?: string[]
  studentReadableDataKeys?: string[]
}

export function getParticipantActions(
  actions: {
    setPassed: (rows: ParticipantRow[]) => void
    setFailed: (rows: ParticipantRow[]) => void
  },
  extraActions: RowAction<ParticipantRow>[],
  exportDeps: ExportDeps,
  extraColumns: ExtraParticipantColumn<any>[] = [],
): RowAction<ParticipantRow>[] {
  return [
    {
      label: 'Set Passed',
      icon: <CheckCircle className='h-4 w-4' />,
      confirm: {
        title: 'Confirm',
        description: (c) => `Mark ${c} participants as passed?`,
        confirmLabel: 'Set Passed',
      },
      onAction: actions.setPassed,
    },
    {
      label: 'Set Failed',
      icon: <XCircle className='h-4 w-4' />,
      confirm: {
        title: 'Confirm',
        description: (c) => `Mark ${c} participants as failed?`,
        confirmLabel: 'Set Failed',
        variant: 'destructive',
      },
      onAction: actions.setFailed,
    },
    ...extraActions,
    {
      label: 'Export CSV',
      icon: <Download className='h-4 w-4' />,
      onAction: (rows) => {
        const originalRows = rows.map((r) => ({
          coursePhaseID: r.coursePhaseID,
          courseParticipationID: r.courseParticipationID,
          passStatus: r.passStatus,
          restrictedData: r.restrictedData,
          studentReadableData: r.studentReadableData ?? {},
          prevData: r.prevData ?? {},
          student: r.student,
        }))

        downloadParticipations(
          originalRows,
          exportDeps.prevDataKeys ?? [],
          exportDeps.restrictedDataKeys ?? [],
          exportDeps.studentReadableDataKeys ?? [],
          extraColumns,
        )
      },
    },
  ]
}
