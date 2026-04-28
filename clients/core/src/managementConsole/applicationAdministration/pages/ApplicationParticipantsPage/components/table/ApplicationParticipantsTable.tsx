import { ReactNode, useCallback, useMemo, useRef } from 'react'
import { PromptTable, TableFilter, useToast } from '@tumaet/prompt-ui-components'

import { useDeleteApplications } from '../../hooks/useDeleteApplications'
import { useApplicationStore } from '@core/managementConsole/applicationAdministration/zustand/useApplicationStore'
import { ApplicationRow, buildApplicationRows } from './applicationRow'
import { getApplicationColumns } from './applicationColumns'
import { getApplicationFilters } from './applicationFilters'
import { getApplicationActions } from './applicationActions'
import { PassStatus } from '@tumaet/prompt-shared-state'
import { useUpdateCoursePhaseParticipationBatch } from '@/hooks/useUpdateCoursePhaseParticipationBatch'
import { ColumnDef } from '@tanstack/react-table'
import { useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { getApplicationForm } from '@core/network/queries/applicationForm'
import { getApplicationAssessment } from '@core/network/queries/applicationAssessment'
import { downloadApplications } from '../../utils/downloadApplications'
import { getApplicationCsvExportSettings } from '@core/managementConsole/applicationAdministration/utils/applicationCsvExportSettings'

const APPLICATION_EXPORT_CONCURRENCY = 10

const fetchWithConcurrencyLimit = async <TItem, TResult>(
  items: TItem[],
  limit: number,
  fetchItem: (item: TItem) => Promise<TResult>,
): Promise<PromiseSettledResult<TResult>[]> => {
  const results = new Array<PromiseSettledResult<TResult>>(items.length)
  let nextIndex = 0

  const workers = Array.from({ length: Math.min(limit, items.length) }, async () => {
    while (nextIndex < items.length) {
      const currentIndex = nextIndex
      nextIndex += 1

      try {
        results[currentIndex] = {
          status: 'fulfilled',
          value: await fetchItem(items[currentIndex]),
        }
      } catch (reason) {
        results[currentIndex] = {
          status: 'rejected',
          reason,
        }
      }
    }
  })

  await Promise.all(workers)

  return results
}

export const ApplicationParticipantsTable = ({ phaseId }: { phaseId: string }): ReactNode => {
  const { courseId } = useParams()
  const { participations, additionalScores, coursePhase } = useApplicationStore()
  const { mutate: deleteApplications } = useDeleteApplications()
  const navigate = useNavigate()
  const tableContainerRef = useRef<HTMLDivElement | null>(null)
  const { toast } = useToast()

  const data = useMemo(
    () => buildApplicationRows(participations, additionalScores),
    [participations, additionalScores],
  )

  const columns: ColumnDef<ApplicationRow>[] = useMemo(
    () => getApplicationColumns(additionalScores),
    [additionalScores],
  )

  const filters: TableFilter[] = useMemo(
    () => getApplicationFilters(additionalScores),
    [additionalScores],
  )

  const getVisibleApplicationIds = useCallback(() => {
    const idsFromVisibleRows = Array.from(
      tableContainerRef.current?.querySelectorAll('[data-application-participation-id]') ?? [],
    )
      .map((element) => element.getAttribute('data-application-participation-id') ?? '')
      .filter((id, index, ids) => Boolean(id) && ids.indexOf(id) === index)

    if (idsFromVisibleRows.length > 0) {
      return idsFromVisibleRows
    }

    return data.map((row) => row.courseParticipationID)
  }, [data])

  const viewApplication = useCallback(
    (row: ApplicationRow) => {
      navigate(
        `/management/course/${courseId}/${phaseId}/participants/${row.courseParticipationID}`,
        {
          state: {
            filteredApplicationIds: getVisibleApplicationIds(),
          },
        },
      )
    },
    [navigate, courseId, phaseId, getVisibleApplicationIds],
  )

  const queryClient = useQueryClient()
  const { mutate: updateBatch } = useUpdateCoursePhaseParticipationBatch()

  const exportApplications = useCallback(
    async (rows: ApplicationRow[]) => {
      try {
        const applicationForm = await queryClient.fetchQuery({
          queryKey: ['application_form', phaseId],
          queryFn: () => getApplicationForm(phaseId),
        })

        const applicationResults = await fetchWithConcurrencyLimit(
          rows,
          APPLICATION_EXPORT_CONCURRENCY,
          (row) =>
            queryClient.fetchQuery({
              queryKey: ['application', phaseId, row.courseParticipationID],
              queryFn: () => getApplicationAssessment(phaseId, row.courseParticipationID),
            }),
        )

        const applicationsByParticipationID = Object.fromEntries(
          rows.flatMap((row, index) => {
            const result = applicationResults[index]
            return result.status === 'fulfilled' ? [[row.courseParticipationID, result.value]] : []
          }),
        )
        const failedCount = applicationResults.filter(
          (result) => result.status === 'rejected',
        ).length

        downloadApplications(
          rows,
          additionalScores,
          'application-export.csv',
          applicationForm,
          applicationsByParticipationID,
          getApplicationCsvExportSettings(coursePhase?.restrictedData),
        )

        if (failedCount > 0) {
          toast({
            title: 'Export completed with missing application answers.',
            description: `${failedCount} application${failedCount === 1 ? '' : 's'} could not be loaded. Base participant data was still exported.`,
            variant: 'destructive',
          })
          return
        }

        toast({
          title: 'Successfully exported applications.',
        })
      } catch {
        toast({
          title: 'Error',
          description: 'Failed to export applications.',
          variant: 'destructive',
        })
      }
    },
    [additionalScores, coursePhase?.restrictedData, phaseId, queryClient, toast],
  )

  const actions = useMemo(() => {
    const setStatus = (status: PassStatus, rows: ApplicationRow[]) => {
      updateBatch(
        rows.map((r) => ({
          coursePhaseID: phaseId,
          courseParticipationID: r.courseParticipationID,
          passStatus: status,
          restrictedData: r.restrictedData ?? {},
          studentReadableData: {},
        })),
        {
          onSuccess: () => {
            queryClient.invalidateQueries({
              queryKey: ['application_participations', 'students', phaseId],
            })
          },
        },
      )
    }
    return getApplicationActions(deleteApplications, viewApplication, {
      setPassed: (r) => setStatus(PassStatus.PASSED, r),
      setFailed: (r) => setStatus(PassStatus.FAILED, r),
      exportCsv: exportApplications,
    })
  }, [deleteApplications, viewApplication, phaseId, updateBatch, queryClient, exportApplications])

  return (
    <div ref={tableContainerRef}>
      <PromptTable<ApplicationRow>
        data={data}
        columns={columns}
        filters={filters}
        actions={actions}
        onRowClick={viewApplication}
        initialState={{ sorting: [{ id: 'firstName', desc: false }] }}
      />
    </div>
  )
}
