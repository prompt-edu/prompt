import {
  ExportStatus,
  type PrivacyExport,
  type PrivacyExportDocument,
} from '@core/network/queries/privacyStudentDataExport'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@tumaet/prompt-ui-components'
import { ChevronDown } from 'lucide-react'
import { useState } from 'react'
import { PrivacyExportDocument as PrivacyExportDocumentCard } from './PrivacyExportDoc'

interface PrivacyExportDocumentListProps {
  privacyExport: PrivacyExport
}

function CollapsibleDocSection({
  title,
  description,
  docs,
  exportId,
}: {
  title: string
  description: string
  docs: PrivacyExportDocument[]
  exportId: string
}) {
  const [open, setOpen] = useState(false)

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className='flex items-center gap-1.5 cursor-pointer py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors'>
        <ChevronDown
          className='h-4 w-4 transition-transform duration-200'
          style={{ transform: open ? 'rotate(180deg)' : 'rotate(0deg)' }}
        />
        <span>
          {title} ({docs.length})
        </span>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <p className='text-muted-foreground pb-2'>{description}</p>
        <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3'>
          {docs.map((doc) => (
            <PrivacyExportDocumentCard
              key={doc.id}
              exportId={exportId}
              privacy_export_document={doc}
            />
          ))}
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

export function PrivacyExportDocumentList({ privacyExport }: PrivacyExportDocumentListProps) {
  const { documents, id: exportId } = privacyExport

  const mainDocs = documents.filter(
    (d) => d.status !== ExportStatus.failed && d.status !== ExportStatus.no_data,
  )
  const failedDocs = documents.filter((d) => d.status === ExportStatus.failed)
  const noDataDocs = documents.filter((d) => d.status === ExportStatus.no_data)

  if (documents.length === 0) return null

  return (
    <div className='space-y-4'>
      {mainDocs.length > 0 && (
        <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3'>
          {mainDocs.map((doc) => (
            <PrivacyExportDocumentCard
              key={doc.id}
              exportId={exportId}
              privacy_export_document={doc}
            />
          ))}
        </div>
      )}

      {(noDataDocs.length > 0 || failedDocs.length > 0) && (
        <div className='space-y-1'>
          {noDataDocs.length > 0 && (
            <CollapsibleDocSection
              title='No data'
              description='The export request was successful. There is no data stored about you on these microservices.'
              docs={noDataDocs}
              exportId={exportId}
            />
          )}
          {failedDocs.length > 0 && (
            <CollapsibleDocSection
              title='Failed'
              description="These microservices returned an error. Please contact an administrator or Prompt's Privacy Contact."
              docs={failedDocs}
              exportId={exportId}
            />
          )}
        </div>
      )}
    </div>
  )
}
