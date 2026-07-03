import type { PrivacyExportDocument } from '@core/network/queries/privacyStudentDataExport'
import { Collapsible, CollapsibleTrigger } from '@tumaet/prompt-ui-components'
import { ChevronDown } from 'lucide-react'
import { useState } from 'react'
import { PrivacyExportDocument as PrivacyExportDocumentCard } from './PrivacyExportDoc'

interface DocSectionProps {
  title: string
  description?: string
  docs: PrivacyExportDocument[]
  exportId: string
  defaultOpen?: boolean
}

export function CollapsibleDocSection({
  title,
  description,
  docs,
  exportId,
  defaultOpen = true,
}: DocSectionProps) {
  const [open, setOpen] = useState(defaultOpen)

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className='flex items-center gap-1.5 cursor-pointer py-2 font-medium text-muted-foreground hover:text-foreground'>
        <ChevronDown className={`h-4 w-4 ${open ? '' : '-rotate-90'}`} />
        <span>
          {title} ({docs.length})
        </span>
      </CollapsibleTrigger>

      {open && (
        <div>
          {description && <p className='text-xs text-muted-foreground pb-2'>{description}</p>}
          <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 pb-1'>
            {docs.map((doc) => (
              <PrivacyExportDocumentCard
                key={doc.id}
                exportId={exportId}
                privacy_export_document={doc}
              />
            ))}
          </div>
        </div>
      )}
    </Collapsible>
  )
}
