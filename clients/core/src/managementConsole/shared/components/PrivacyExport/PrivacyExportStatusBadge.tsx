import { ExportStatus } from '@core/network/queries/privacyStudentDataExport'
import { AnimatePresence, motion } from 'framer-motion'
import { CircleCheck, CircleDashed, CircleX } from 'lucide-react'

interface PrivacyExportStatusProps {
  privacy_export_status: ExportStatus
  size?: number
}

const statusStyles: Record<ExportStatus, string> = {
  [ExportStatus.failed]: 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400',
  [ExportStatus.pending]: 'bg-muted text-muted-foreground',
  [ExportStatus.complete]: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
}

const statusIcons = (size: number): Record<ExportStatus, React.ReactNode> => ({
  [ExportStatus.failed]: <CircleX size={size} />,
  [ExportStatus.pending]: <CircleDashed className='animate-spin' size={size} />,
  [ExportStatus.complete]: <CircleCheck size={size} />,
})

export function PrivacyExportStatus({
  privacy_export_status,
  size = 20,
}: PrivacyExportStatusProps) {
  return (
    <div className={'inline-flex shrink-0 p-2 rounded-full ' + statusStyles[privacy_export_status]}>
      <AnimatePresence mode='wait' initial={false}>
        <motion.span
          key={privacy_export_status}
          className='flex'
          initial={{ scale: 0.5 }}
          animate={{ scale: 1 }}
          transition={{ type: 'spring', stiffness: 400, damping: 20 }}
        >
          {statusIcons(size)[privacy_export_status]}
        </motion.span>
      </AnimatePresence>
    </div>
  )
}
