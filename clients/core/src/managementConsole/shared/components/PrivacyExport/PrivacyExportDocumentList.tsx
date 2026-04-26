import { ExportStatus, type PrivacyExport } from '@core/network/queries/privacyStudentDataExport'
import { AnimatePresence, motion } from 'framer-motion'
import { AnimatedDocSection } from './PrivacyExportDocSection'

interface PrivacyExportDocumentListProps {
  privacyExport: PrivacyExport
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
    <motion.div layout className='space-y-1'>
      <AnimatePresence initial={false}>
        {mainDocs.length > 0 && (
          <AnimatedDocSection title='Ready' docs={mainDocs} exportId={exportId} />
        )}

        {noDataDocs.length > 0 && (
          <AnimatedDocSection
            title='No data'
            description='The export request was successful. There is no data stored about you on these microservices.'
            docs={noDataDocs}
            exportId={exportId}
          />
        )}

        {failedDocs.length > 0 && (
          <AnimatedDocSection
            title='Failed'
            description='There was a problem exporting these parts. Please contact an administrator or Prompts Privacy Contact.'
            docs={failedDocs}
            exportId={exportId}
          />
        )}
      </AnimatePresence>
    </motion.div>
  )
}
