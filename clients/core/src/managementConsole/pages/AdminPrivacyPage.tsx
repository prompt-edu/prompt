import {
  ExportStatus,
  getAllExports,
  type AdminPrivacyExport,
} from '@core/network/queries/privacyStudentDataExport'
import { ManagementPageHeader, PromptTable } from '@tumaet/prompt-ui-components'
import { useEffect, useState } from 'react'
import { adminExportColumns } from '../shared/components/PrivacyExport/adminExportColumns'

export function AdminPrivacyPage() {
  const [exports, setExports] = useState<AdminPrivacyExport[]>([])

  useEffect(() => {
    getAllExports().then(setExports)
  }, [])

  return (
    <div>
      <ManagementPageHeader>Privacy Dashboard</ManagementPageHeader>
      <p className='mb-8 text-muted-foreground'>
        Monitor and manage privacy-related requests from users.
      </p>

      <div className='space-y-10'>
        <section>
          <h2 className='text-lg font-semibold text-foreground mb-4'>Deletion Requests</h2>
          <PromptTable data={[]} columns={[]} pageSize={10} />
        </section>

        <section>
          <h2 className='text-lg font-semibold text-foreground mb-4'>Data Exports</h2>
          <PromptTable
            data={exports}
            columns={adminExportColumns}
            filters={[
              {
                type: 'select',
                id: 'status',
                label: 'Status',
                options: Object.values(ExportStatus),
              },
            ]}
            pageSize={10}
          />
        </section>
      </div>
    </div>
  )
}
