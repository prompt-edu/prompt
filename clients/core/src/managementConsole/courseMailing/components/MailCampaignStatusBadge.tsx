import { Badge } from '@tumaet/prompt-ui-components'
import { MailCampaignStatus } from '../interfaces/mailCampaign'

const statusConfig: Record<MailCampaignStatus, { label: string; className: string }> = {
  [MailCampaignStatus.Draft]: {
    label: 'Draft',
    className: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  },
  [MailCampaignStatus.Sending]: {
    label: 'Sending',
    className: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  },
  [MailCampaignStatus.Sent]: {
    label: 'Sent',
    className: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  },
  [MailCampaignStatus.PartiallyFailed]: {
    label: 'Partially failed',
    className: 'bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300',
  },
  [MailCampaignStatus.Failed]: {
    label: 'Failed',
    className: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
  },
}

interface MailCampaignStatusBadgeProps {
  status: MailCampaignStatus
}

export const MailCampaignStatusBadge = ({ status }: MailCampaignStatusBadgeProps) => {
  const config = statusConfig[status] ?? statusConfig[MailCampaignStatus.Draft]
  return (
    <Badge variant='outline' className={config.className}>
      {config.label}
    </Badge>
  )
}
