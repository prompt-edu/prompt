import { ExportStatus } from '@core/network/queries/privacyStudentDataExport'
import { AnimatePresence, motion } from 'framer-motion'
import { CircleCheck, CircleDashed, CircleX } from 'lucide-react'

interface PrivacyExportStatusProps {
  privacy_export_status: ExportStatus
}

const statusStyles: Record<ExportStatus, string> = {
  [ExportStatus.failed]: 'bg-red-200 text-red-800',
  [ExportStatus.pending]: 'bg-gray-200 text-gray-800',
  [ExportStatus.complete]: 'bg-green-200 text-green-800',
}

const statusIcons: Record<ExportStatus, React.ReactNode> = {
  [ExportStatus.failed]: <CircleX size={12} />,
  [ExportStatus.pending]: <CircleDashed className='animate-spin' size={12} />,
  [ExportStatus.complete]: <CircleCheck size={12} />,
}

export function PrivacyExportStatus({ privacy_export_status }: PrivacyExportStatusProps) {
  return (
    <div
      className={
        'px-2 py-1 rounded-full text-xs flex items-center gap-1 ' +
        statusStyles[privacy_export_status]
      }
    >
      <AnimatePresence mode='wait' initial={false}>
        <motion.span
          key={privacy_export_status}
          initial={{ scale: 0.5 }}
          animate={{ scale: 1 }}
          transition={{ type: 'spring', stiffness: 400, damping: 20 }}
        >
          {statusIcons[privacy_export_status]}
        </motion.span>
      </AnimatePresence>
      <span>{privacy_export_status}</span>
    </div>
  )
}
