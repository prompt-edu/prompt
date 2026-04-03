import { type PrivacyExportDocument } from '@core/network/queries/privacyStudentDataExport'
import { AnimatePresence, motion } from 'framer-motion'
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

export function AnimatedDocSection({
  title,
  description,
  docs,
  exportId,
  defaultOpen = true,
}: DocSectionProps) {
  return (
    <motion.div
      key={title}
      layout
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.4 }}
    >
      <CollapsibleDocSection
        title={title}
        description={description}
        docs={docs}
        exportId={exportId}
        defaultOpen={defaultOpen}
      />
    </motion.div>
  )
}
