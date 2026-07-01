import { PassStatus } from '@tumaet/prompt-shared-state'
import { Badge } from '@tumaet/prompt-ui-components'

export function getStatusBadge(status: PassStatus) {
  switch (status) {
    case 'passed':
      return <Badge className='bg-green-500 hover:bg-green-500'>Accepted</Badge>
    case 'failed':
      return <Badge className='bg-red-500 hover:bg-red-500'>Rejected</Badge>
    case 'not_assessed':
      return <Badge className='bg-gray-500 hover:bg-gray-500'>Not Assessed</Badge>
    default:
      return <Badge className='bg-gray-500 hover:bg-gray-500'>Unknown</Badge>
  }
}

export function getStatusString(status: PassStatus): string {
  switch (status) {
    case 'passed':
      return 'Accepted'
    case 'failed':
      return 'Rejected'
    case 'not_assessed':
      return 'Not Assessed'
    default:
      return 'Unknown'
  }
}
