import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { useTriggerAssessment } from '../pages/hooks/useTriggerAssessment'
import { useGetDeveloperProfile } from '../pages/hooks/useGetDeveloperProfile'
import { getOwnCoursePhaseParticipation } from '@/network/queries/getOwnCoursePhaseParticipation'
import {
  Button,
  Alert,
  AlertDescription,
  AlertTitle,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Badge,
  Checkbox,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  cn,
} from '@tumaet/prompt-ui-components'
import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import {
  RefreshCw,
  CheckCircle,
  AlertCircle,
  Hourglass,
  FileCode2,
  Clock,
  PartyPopper,
  CircleX,
  Loader2,
  SearchCode,
  AlertTriangle,
} from 'lucide-react'

export const Assessment = () => {
  const [error, setError] = useState<string | null>(null)
  const [confirmedOwnWork, setConfirmedOwnWork] = useState(false)
  const [showConfirmDialog, setShowConfirmDialog] = useState(false)
  const [assessmentTriggered, setAssessmentTriggered] = useState(false)

  const assessmentMutation = useTriggerAssessment(setError)
  const developerQuery = useGetDeveloperProfile()

  const remainingAttempts = Math.max(
    (developerQuery.data?.maxAttempts ?? 0) - (developerQuery.data?.attempts ?? 0),
    0,
  )
  const maxAttempts = developerQuery.data?.maxAttempts ?? 0
  const passingPosition = developerQuery.data?.position ?? undefined
  const passed = developerQuery.data?.hasPassed

  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: fetchedParticipation,
    isError: participationError,
    isLoading: isParticipationPending,
  } = useQuery<CoursePhaseParticipationWithStudent>({
    queryKey: ['course_phase_participation', phaseId],
    queryFn: () => getOwnCoursePhaseParticipation(phaseId ?? ''),
  })

  const handleTriggerAssessment = () => {
    assessmentMutation.mutate(undefined, {
      onSettled: () => {
        setAssessmentTriggered(true)
      },
    })
    setShowConfirmDialog(false)
  }

  const openConfirmDialog = () => {
    setShowConfirmDialog(true)
  }

  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-xl flex items-center'>
          <FileCode2 className='mr-2 h-5 w-5' />
          Assessment
          <div className='flex items-center space-x-2 ml-auto'>
            <Badge
              variant='outline'
              className={cn('text-sm', {
                'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300':
                  remainingAttempts === maxAttempts || !assessmentTriggered,
                'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300': passed,
                'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300':
                  !passed && (remainingAttempts !== maxAttempts || assessmentTriggered),
              })}
            >
              {passed
                ? 'Passed'
                : !passed && (remainingAttempts !== maxAttempts || assessmentTriggered)
                  ? 'Not Passed'
                  : !passed && remainingAttempts === 0
                    ? 'Failed'
                    : 'Not Started'}
            </Badge>
            <Badge
              variant='outline'
              className={cn('text-sm', {
                'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300':
                  remainingAttempts >= 3,
                'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300':
                  remainingAttempts === 2,
                'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300': remainingAttempts < 2,
              })}
            >
              Remaining Attempts: {remainingAttempts}
            </Badge>
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className='flex items-start space-x-2 mt-4'>
          {isParticipationPending && (
            <div className='flex justify-center items-center h-64'>
              <Loader2 className='h-12 w-12 animate-spin text-primary' />
            </div>
          )}
          {participationError && (
            <Alert variant='destructive'>
              <AlertCircle className='h-4 w-4' />
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>
                An error occurred while fetching the assessment status. Please try again later.
              </AlertDescription>
            </Alert>
          )}
          {passed &&
            ((passingPosition !== undefined && passingPosition <= 150) ||
              fetchedParticipation?.passStatus === 'passed') && (
              <Alert variant='default'>
                <PartyPopper className='h-4 w-4' />
                <AlertTitle>Congratulations</AlertTitle>
                <AlertDescription>
                  You have successfully completed the technical challenge and are eligible to
                  participate in the DevOps course!
                </AlertDescription>
              </Alert>
            )}
          {passed &&
            passingPosition !== undefined &&
            passingPosition > 150 &&
            fetchedParticipation?.passStatus !== 'passed' && (
              <Alert variant='default'>
                <Clock className='h-4 w-4' />
                <AlertTitle>Waitlisted</AlertTitle>
                <AlertDescription>
                  You have successfully completed the technical challenge, but due to limited
                  capacity, you have been placed on the waitlist for the DevOps course.
                </AlertDescription>
              </Alert>
            )}
          {!passed && remainingAttempts === 0 && (
            <Alert variant='destructive'>
              <CircleX className='h-4 w-4' />
              <AlertTitle>Challenge Unsuccessful</AlertTitle>
              <AlertDescription>
                You did not successfully complete the technical challenge and have no remaining
                attempts.
              </AlertDescription>
            </Alert>
          )}
        </div>
        <div className='mt-4'>
          <div className='flex items-center mb-2'>
            <h3 className='text-lg font-medium'>Last Assessment Result</h3>
            {remainingAttempts !== maxAttempts || assessmentTriggered ? (
              <span className='ml-2'>
                {passed ? (
                  <CheckCircle className='h-5 w-5 text-green-500' />
                ) : (
                  <AlertCircle className='h-5 w-5 text-red-500' />
                )}
              </span>
            ) : (
              <span className='ml-2'>
                <Hourglass className='h-5 w-5 text-gray-500' />
              </span>
            )}
          </div>

          <div className='space-y-3 mt-2'>
            {remainingAttempts !== maxAttempts || assessmentTriggered ? (
              <Alert variant={!passed ? 'destructive' : 'default'}>
                <div className='flex items-center'>
                  {passed ? (
                    <CheckCircle className='h-4 w-4 text-green-500' />
                  ) : (
                    <AlertCircle className='h-4 w-4' />
                  )}
                  <AlertTitle className={`ml-2 ${passed ? 'text-green-500' : ''}`}>
                    {!passed ? 'Error' : 'Success'}
                  </AlertTitle>
                </div>
                <AlertDescription className={`mt-1 ${passed ? 'text-green-500' : ''}`}>
                  {!passed
                    ? (error ?? 'Your last testing attempt did not pass.')
                    : 'You passed the challenge!'}
                </AlertDescription>
              </Alert>
            ) : (
              <Alert variant='default'>
                <Hourglass className='h-4 w-4' />
                <AlertTitle>Test Not Started</AlertTitle>
                <AlertDescription>You have not yet attempted the challenge.</AlertDescription>
              </Alert>
            )}
          </div>
        </div>

        <div className='flex items-start space-x-2 mt-4'>
          <Checkbox
            id='own-work-check'
            checked={passed || remainingAttempts === 0 || confirmedOwnWork}
            onCheckedChange={(checked) => !passed && setConfirmedOwnWork(checked as boolean)}
            disabled={passed || remainingAttempts === 0}
          />
          <label
            htmlFor='own-work-check'
            className={`text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer ${
              passed || remainingAttempts === 0 ? 'opacity-70 cursor-not-allowed' : ''
            }`}
          >
            I confirm that I have completed this challenge independently.
          </label>
        </div>

        <div className='flex items-start space-x-2 mt-4'>
          <Button
            onClick={openConfirmDialog}
            disabled={
              assessmentMutation.isPending || remainingAttempts === 0 || !confirmedOwnWork || passed
            }
            className='w-full'
          >
            {assessmentMutation.isPending ? (
              <RefreshCw className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <SearchCode className='mr-2 h-4 w-4' />
            )}
            <span>{assessmentMutation.isPending ? 'Testing...' : 'Start Testing'}</span>
          </Button>
        </div>

        <Dialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle className='flex items-center'>
                <AlertTriangle className='h-5 w-5 text-amber-500 mr-2' />
                Confirm Assessment Attempt
              </DialogTitle>
              <DialogDescription className='pt-2'>
                Starting this test will decrease your remaining attempts by 1. You currently have{' '}
                <strong>{remainingAttempts}</strong> attempt{remainingAttempts !== 1 ? 's' : ''}{' '}
                remaining.
                <Alert variant='default' className='mt-4'>
                  <AlertTriangle className='h-4 w-4' />
                  <AlertTitle>Important</AlertTitle>
                  <AlertDescription>
                    Once you start the test, the attempt will be counted even if you close the
                    browser or encounter technical issues.
                  </AlertDescription>
                </Alert>
              </DialogDescription>
            </DialogHeader>
            <DialogFooter className='flex space-x-2 sm:space-x-0'>
              <Button variant='outline' onClick={() => setShowConfirmDialog(false)}>
                Cancel
              </Button>
              <Button onClick={handleTriggerAssessment} variant='default'>
                Proceed with Testing
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  )
}
