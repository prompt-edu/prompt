export enum ChallengeStatus {
  PASSED = 'passed',
  NOT_COMPLETED = 'notCompleted',
  NOT_STARTED = 'notStarted',
}
export function getChallengeStatusString(status: ChallengeStatus | undefined): string {
  switch (status) {
    case ChallengeStatus.PASSED:
      return 'passed'
    case ChallengeStatus.NOT_COMPLETED:
      return 'notCompleted'
    default:
      return 'notStarted'
  }
}
