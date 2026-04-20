import { ChallengeStatus } from '../interfaces/challengeStatus'
import { Badge } from '@tumaet/prompt-ui-components'

export function getChallengeStatusBadgeFromString(status: string) {
  switch (status) {
    case 'passed':
      return (
        <Badge className='bg-green-100 border border-green-500 text-green-700 hover:bg-green-100 hover:text-green-700'>
          Passed
        </Badge>
      )
    case 'notCompleted':
      return (
        <Badge className='bg-red-100 border border-red-500 text-red-700 hover:bg-red-100 hover:text-red-700'>
          Not Completed
        </Badge>
      )
    default:
      return (
        <Badge className='bg-gray-100 border border-gray-500 text-gray-700 hover:bg-gray-100 hover:text-gray-700'>
          Not Started
        </Badge>
      )
  }
}

export function getChallengeStatusBadge(challengeStatus: ChallengeStatus | undefined) {
  if (challengeStatus == ChallengeStatus.PASSED) {
    return getChallengeStatusBadgeFromString('passed')
  } else if (challengeStatus == ChallengeStatus.NOT_COMPLETED) {
    return getChallengeStatusBadgeFromString('notCompleted')
  } else {
    return getChallengeStatusBadgeFromString('')
  }
}
