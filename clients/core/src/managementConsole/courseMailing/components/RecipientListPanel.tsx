import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@tumaet/prompt-ui-components'
import type { RecipientPreviewItem } from '../interfaces/mailCampaign'

interface RecipientListPanelProps {
  recipients: RecipientPreviewItem[]
  isLoading?: boolean
}

export const RecipientListPanel = ({ recipients, isLoading }: RecipientListPanelProps) => {
  if (isLoading) {
    return <p className='text-sm text-muted-foreground'>Loading recipients...</p>
  }

  if (recipients.length === 0) {
    return (
      <p className='text-sm text-muted-foreground'>
        No recipients match the selected phase and statuses.
      </p>
    )
  }

  return (
    <div className='max-h-72 overflow-y-auto rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {recipients.map((recipient) => (
            <TableRow key={recipient.courseParticipationID}>
              <TableCell>
                {recipient.firstName} {recipient.lastName}
              </TableCell>
              <TableCell>
                {recipient.email || <span className='text-red-500'>no email</span>}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
