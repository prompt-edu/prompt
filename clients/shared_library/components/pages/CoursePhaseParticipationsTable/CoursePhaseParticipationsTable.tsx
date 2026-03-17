import { useMemo } from 'react'
import {
  buildParticipantRows,
  ExtraParticipantColumn,
  ParticipantRow,
} from './table/participationRow'
import { getParticipantColumns } from './table/participationColumns'
import { getParticipantFilters } from './table/participationFilters'
import { ExportDeps, getParticipantActions } from './table/participationActions'
import { PromptTable, RowAction, TableFilter } from '@tumaet/prompt-ui-components'
import { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { useParticipantBatchActions } from './utils/updateBatch'

interface CoursePhaseParticipationsTablePageProps {
  phaseId: string
  participants: CoursePhaseParticipationWithStudent[]
  extraColumns?: ExtraParticipantColumn<any>[]
  extraFilters?: TableFilter[]
  extraActions?: RowAction<ParticipantRow>[]
  exportDeps?: ExportDeps
  onClickRowAction?: (row: ParticipantRow) => void
}

export const CoursePhaseParticipationsTable = ({
  phaseId,
  participants,
  extraColumns = [],
  extraFilters = [],
  extraActions = [],
  exportDeps = {},
  onClickRowAction,
}: CoursePhaseParticipationsTablePageProps) => {
  const data = useMemo(
    () => buildParticipantRows(participants, extraColumns),
    [participants, extraColumns],
  )

  const columns = useMemo(() => getParticipantColumns(extraColumns), [extraColumns])

  const filters = useMemo(() => getParticipantFilters(extraFilters), [extraFilters])

  const { setPassed, setFailed } = useParticipantBatchActions(phaseId)
  const actions = useMemo(
    () =>
      getParticipantActions({ setPassed, setFailed }, extraActions, exportDeps ?? {}, extraColumns),
    [setPassed, setFailed, extraActions, exportDeps, extraColumns],
  )

  return (
    <PromptTable<ParticipantRow>
      data={data}
      columns={columns}
      filters={filters}
      actions={actions}
      onRowClick={(row) => onClickRowAction?.(row)}
      initialState={{ columnVisibility: { matriculationNumber: false, universityLogin: false } }}
    />
  )
}
