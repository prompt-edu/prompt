import { getApplicationCsvExportSettings } from '@core/managementConsole/applicationAdministration/utils/applicationCsvExportSettings'
import { useApplicationStore } from '@core/managementConsole/applicationAdministration/zustand/useApplicationStore'
import { getApplicationAssessment } from '@core/network/queries/applicationAssessment'
import { getApplicationForm } from '@core/network/queries/applicationForm'
import { getExportedApplicationAnswers } from '@core/network/queries/exportedApplicationAnswers'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { PassStatus, useUpdateCoursePhaseParticipationBatch } from '@tumaet/prompt-shared-state'
import { PromptTableURL, type TableFilter, useToast } from '@tumaet/prompt-ui-components'
import { type ReactNode, useCallback, useEffect, useMemo, useRef } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useDeleteApplications } from '../../hooks/useDeleteApplications'
import { downloadApplications } from '../../utils/downloadApplications'
import { getApplicationActions } from './applicationActions'
import { getApplicationColumns } from './applicationColumns'
import { getApplicationFilters } from './applicationFilters'
import { type ApplicationRow, buildApplicationRows } from './applicationRow'

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

  const { data: exportedAnswers, isError: exportedAnswersError } = useQuery({
    queryKey: ['application_exported_answers', phaseId],
    queryFn: () => getExportedApplicationAnswers(phaseId),
  })

  useEffect(() => {
    if (exportedAnswersError) {
      toast({
        title: 'Could not load exported application answers.',
        description: 'The participants table is shown without the exported question columns.',
        variant: 'destructive',
      })
    }
  }, [exportedAnswersError, toast])

  const exportedColumns = useMemo(() => exportedAnswers?.columns ?? [], [exportedAnswers])

  const exportedAnswersByParticipation = useMemo(() => {
    const byParticipation = new Map<string, Map<string, string>>()
    for (const participation of exportedAnswers?.answers ?? []) {
      byParticipation.set(
        participation.courseParticipationID,
        new Map(participation.answers.map((answer) => [answer.questionID, answer.answer])),
      )
    }
    return byParticipation
  }, [exportedAnswers])

  const data = useMemo(
    () => buildApplicationRows(participations, additionalScores, exportedAnswersByParticipation),
    [participations, additionalScores, exportedAnswersByParticipation],
  )

  const columns: ColumnDef<ApplicationRow>[] = useMemo(
    () => getApplicationColumns(additionalScores, exportedColumns),
    [additionalScores, exportedColumns],
  )

  const studyPrograms = useMemo(
    () =>
      Array.from(new Set(data.map((row) => row.studyProgram).filter(Boolean))).sort((a, b) =>
        a.localeCompare(b),
      ),
    [data],
  )

  const filters: TableFilter[] = useMemo(
    () => getApplicationFilters(additionalScores, studyPrograms),
    [additionalScores, studyPrograms],
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
      <PromptTableURL<ApplicationRow>
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
