import {
  ExportStatus,
  type PrivacyExport,
  type PrivacyExportDocument,
} from '@core/network/queries/privacyStudentDataExport'
import { Collapsible, CollapsibleTrigger } from '@tumaet/prompt-ui-components'
import { AnimatePresence, motion } from 'framer-motion'
import { ChevronDown } from 'lucide-react'
import { useState } from 'react'
import { PrivacyExportDocument as PrivacyExportDocumentCard } from './PrivacyExportDoc'

interface PrivacyExportDocumentListProps {
  privacyExport: PrivacyExport
  inProgress: boolean
}

function CollapsibleDocSection({
  title,
  description,
  docs,
  exportId,
  defaultOpen = false,
}: {
  title: string
  description?: string
  docs: PrivacyExportDocument[]
  exportId: string
  defaultOpen?: boolean
}) {
  const [open, setOpen] = useState(defaultOpen)

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className='flex items-center gap-1.5 cursor-pointer py-2 font-medium text-muted-foreground hover:text-foreground transition-colors'>
        <ChevronDown
          className='h-4 w-4 transition-transform duration-200'
          style={{ transform: open ? 'rotate(0deg)' : 'rotate(-90deg)' }}
        />
        <span>
          {title} ({docs.length})
        </span>
      </CollapsibleTrigger>

      {open && (
        <div>
          {description && <p className='text-xs text-muted-foreground pb-2'>{description}</p>}
          <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 pb-1'>
            <AnimatePresence mode='popLayout'>
              {docs.map((doc) => (
                <motion.div
                  key={doc.id}
                  layout
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  transition={{ duration: 0.4 }}
                >
                  <PrivacyExportDocumentCard exportId={exportId} privacy_export_document={doc} />
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        </div>
      )}
    </Collapsible>
  )
}

export function PrivacyExportDocumentList({
  privacyExport,
  inProgress,
}: PrivacyExportDocumentListProps) {
  const { documents, id: exportId } = privacyExport

  const mainDocs = documents.filter(
    (d) => d.status !== ExportStatus.failed && d.status !== ExportStatus.no_data,
  )
  const failedDocs = documents.filter((d) => d.status === ExportStatus.failed)
  const noDataDocs = documents.filter((d) => d.status === ExportStatus.no_data)

  if (documents.length === 0) return null

  return (
    <motion.div layout className='space-y-1'>
      <AnimatePresence initial={false}>
        {mainDocs.length > 0 && (
          <motion.div
            key='main'
            layout
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.4 }}
          >
            <CollapsibleDocSection
              title='Documents'
              docs={mainDocs}
              exportId={exportId}
              defaultOpen={true}
            />
          </motion.div>
        )}

        {noDataDocs.length > 0 && (
          <motion.div
            key='no-data'
            layout
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.4 }}
          >
            <CollapsibleDocSection
              title='No data'
              description='The export request was successful. There is no data stored about you on these microservices.'
              docs={noDataDocs}
              exportId={exportId}
              defaultOpen={inProgress}
            />
          </motion.div>
        )}

        {failedDocs.length > 0 && (
          <motion.div
            key='failed'
            layout
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.4 }}
          >
            <CollapsibleDocSection
              title='Failed'
              description="These microservices returned an error. Please contact an administrator or Prompt's Privacy Contact."
              docs={failedDocs}
              exportId={exportId}
              defaultOpen={true}
            />
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  )
}
