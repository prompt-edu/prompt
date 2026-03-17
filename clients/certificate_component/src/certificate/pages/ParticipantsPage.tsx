import { useCallback, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Download, Loader2, CheckCircle2, XCircle } from 'lucide-react'

import { ManagementPageHeader, ErrorPage } from '@tumaet/prompt-ui-components'
import { CoursePhaseParticipationsTable } from '@/components/pages/CoursePhaseParticipationsTable/CoursePhaseParticipationsTable'
import {
  ExtraParticipantColumn,
  ParticipantRow,
} from '@/components/pages/CoursePhaseParticipationsTable/table/participationRow'
import { RowAction } from '@tumaet/prompt-ui-components'

import { ParticipantWithDownloadStatus } from '../interfaces/participant'
import { getParticipants } from '../network/queries/getParticipants'
import {
  downloadStudentCertificate,
  triggerBlobDownload,
} from '../network/queries/downloadCertificate'

export const ParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  const {
    data: participants,
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['participants', phaseId],
    queryFn: () => getParticipants(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const handleDownload = useCallback(
    (studentId: string, lastName: string) => {
      if (!phaseId) return

      downloadStudentCertificate(phaseId, studentId)
        .then((blob) => {
          triggerBlobDownload(blob, `certificate_${lastName}.pdf`)
          queryClient.invalidateQueries({ queryKey: ['participants', phaseId] })
        })
        .catch((error) => {
          console.error('Failed to download certificate:', error)
        })
    },
    [phaseId, queryClient],
  )

  // Build a lookup map of download data by student ID
  const downloadDataMap = useMemo(() => {
    const map = new Map<string, ParticipantWithDownloadStatus>()
    for (const p of participants ?? []) {
      map.set(p.courseParticipationID, p)
    }
    return map
  }, [participants])

  const extraColumns: ExtraParticipantColumn<any>[] = useMemo(
    () => [
      {
        id: 'downloadStatus',
        header: 'Download Status',
        extraData: (participants ?? []).map((p) => ({
          courseParticipationID: p.courseParticipationID,
          value: p.hasDownloaded,
          stringValue: p.hasDownloaded ? 'Downloaded' : 'Not downloaded',
        })),
        cell: ({ row }: { row: { original: ParticipantRow } }) => {
          const p = downloadDataMap.get(row.original.courseParticipationID)
          const hasDownloaded = p?.hasDownloaded ?? false
          return (
            <div className='flex items-center gap-2'>
              {hasDownloaded ? (
                <>
                  <CheckCircle2 className='h-4 w-4 text-green-500' />
                  <span className='text-green-600'>Downloaded</span>
                </>
              ) : (
                <>
                  <XCircle className='h-4 w-4 text-muted-foreground' />
                  <span className='text-muted-foreground'>Not downloaded</span>
                </>
              )}
            </div>
          )
        },
      },
      {
        id: 'downloadCount',
        header: 'Downloads',
        extraData: (participants ?? []).map((p) => ({
          courseParticipationID: p.courseParticipationID,
          value: p.downloadCount,
          stringValue: String(p.downloadCount ?? 0),
        })),
        cell: ({ row }: { row: { original: ParticipantRow } }) => {
          const p = downloadDataMap.get(row.original.courseParticipationID)
          return p?.downloadCount ?? 0
        },
      },
      {
        id: 'lastDownload',
        header: 'Last Download',
        extraData: (participants ?? []).map((p) => ({
          courseParticipationID: p.courseParticipationID,
          value: p.lastDownload,
          stringValue: p.lastDownload ? new Date(p.lastDownload).toLocaleDateString() : '-',
        })),
        cell: ({ row }: { row: { original: ParticipantRow } }) => {
          const p = downloadDataMap.get(row.original.courseParticipationID)
          if (!p?.lastDownload) return <span className='text-muted-foreground'>-</span>
          return new Date(p.lastDownload).toLocaleDateString()
        },
      },
    ],
    [participants, downloadDataMap],
  )

  const extraActions: RowAction<ParticipantRow>[] = useMemo(
    () => [
      {
        label: 'Download Certificate',
        icon: <Download className='h-4 w-4' />,
        onAction: (rows) => {
          for (const row of rows) {
            if (row.student?.id) {
              handleDownload(row.student.id, row.lastName)
            }
          }
        },
      },
    ],
    [handleDownload],
  )

  if (isError) {
    return <ErrorPage message='Error loading participants' onRetry={refetch} />
  }

  if (isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  const downloadedCount = participants?.filter((p) => p.hasDownloaded).length ?? 0
  const totalCount = participants?.length ?? 0

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>Certificate Participants</ManagementPageHeader>
      <p className='text-muted-foreground'>View and download certificates for all participants.</p>

      <div className='p-4 bg-muted rounded-lg'>
        <p className='text-sm'>
          <span className='font-medium'>{downloadedCount}</span> of{' '}
          <span className='font-medium'>{totalCount}</span> participants have downloaded their
          certificate.
        </p>
      </div>

      <CoursePhaseParticipationsTable
        phaseId={phaseId!}
        participants={participants ?? []}
        extraColumns={extraColumns}
        extraActions={extraActions}
      />
    </div>
  )
}
