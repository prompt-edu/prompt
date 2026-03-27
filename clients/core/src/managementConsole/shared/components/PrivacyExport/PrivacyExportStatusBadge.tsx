import { ExportStatus } from '@core/network/queries/privacyStudentDataExport'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'
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
  [ExportStatus.no_data]: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
}

const statusIcons = (size: number): Record<ExportStatus, React.ReactNode> => ({
  [ExportStatus.failed]: <CircleX size={size} />,
  [ExportStatus.pending]: <CircleDashed className='animate-spin' size={size} />,
  [ExportStatus.complete]: <CircleCheck size={size} />,
  [ExportStatus.no_data]: <CircleCheck size={size} />,
})

const statusDescription: Record<ExportStatus, string> = {
  [ExportStatus.failed]:
    "The microservice returned with an error. Please contact an administrator or Prompt's Privacy Contact ",
  [ExportStatus.pending]: 'The Export is currently ongoing',
  [ExportStatus.complete]: 'The Export finished succesfully and can be downloaded',
  [ExportStatus.no_data]:
    'The Export request was successful. There is no data stored about you or related to you on this microservice',
}

export function PrivacyExportStatus({
  privacy_export_status,
  size = 20,
}: PrivacyExportStatusProps) {
  return (
    <AnimatePresence mode='wait' initial={false}>
      <motion.div
        key={privacy_export_status}
        className={'inline-flex shrink-0 p-2 rounded-full ' + statusStyles[privacy_export_status]}
        initial={{ scale: 0.5 }}
        animate={{ scale: 1 }}
        transition={{ type: 'spring', stiffness: 400, damping: 20 }}
      >
        <Tooltip>
          <TooltipTrigger className='flex'>
            {statusIcons(size)[privacy_export_status]}
          </TooltipTrigger>
          <TooltipContent>
            <p className='max-w-80'>{statusDescription[privacy_export_status]}</p>
          </TooltipContent>
        </Tooltip>
      </motion.div>
    </AnimatePresence>
  )
}
