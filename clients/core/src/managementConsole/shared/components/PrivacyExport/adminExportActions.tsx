import {
  type AdminPrivacyExport,
  deleteExport,
  ExportStatus,
} from '@core/network/queries/privacyStudentDataExport'
import type { QueryClient } from '@tanstack/react-query'
import type { RowAction } from '@tumaet/prompt-ui-components'
import { Trash2 } from 'lucide-react'

interface PassedDeps {
  queryClient: QueryClient
}

export function getAdminExportActions({
  queryClient,
}: PassedDeps): RowAction<AdminPrivacyExport>[] {
  const invalidateExports = () =>
    queryClient.invalidateQueries({ queryKey: ['privacy', 'admin', 'exports'] })

  return [
    {
      label: 'Delete',
      icon: <Trash2 className='w-4 h-4' />,
      onAction: async (rows) => {
        await Promise.allSettled(rows.map((r) => deleteExport(r.id)))
        invalidateExports()
      },
      hide: (rows) => rows.every((r) => r.status === ExportStatus.archived),
    },
    {
      label: 'Delete + reset rate limit',
      icon: <Trash2 className='w-4 h-4' />,
      onAction: async (rows) => {
        await Promise.allSettled(rows.map((r) => deleteExport(r.id, { resetRateLimit: true })))
        invalidateExports()
      },
    },
  ]
}
