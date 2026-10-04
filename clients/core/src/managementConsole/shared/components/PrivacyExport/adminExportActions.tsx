import { type AdminPrivacyExport, ExportStatus } from '@core/interfaces/privacy'
import { coreApi } from '@core/network/api'
import { coreCache } from '@core/network/cache'
import type { QueryClient } from '@tanstack/react-query'
import type { RowAction } from '@tumaet/prompt-ui-components'
import { Archive } from 'lucide-react'

interface PassedDeps {
  queryClient: QueryClient
}

export function getAdminExportActions({
  queryClient,
}: PassedDeps): RowAction<AdminPrivacyExport>[] {
  const invalidateExports = () => coreCache.privacyExportsChanged(queryClient)

  return [
    {
      label: 'Archive',
      icon: <Archive className='w-4 h-4' />,
      onAction: async (rows) => {
        await Promise.allSettled(rows.map((r) => coreApi.privacy.removeExport(r.id)))
        invalidateExports()
      },
      hide: (rows) => rows.every((r) => r.status === ExportStatus.archived),
    },
    {
      label: 'Archive + reset rate limit',
      icon: <Archive className='w-4 h-4' />,
      onAction: async (rows) => {
        await Promise.allSettled(
          rows.map((r) => coreApi.privacy.removeExport(r.id, { resetRateLimit: true })),
        )
        invalidateExports()
      },
    },
  ]
}
