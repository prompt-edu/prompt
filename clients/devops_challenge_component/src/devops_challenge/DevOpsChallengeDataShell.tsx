import { useParams } from 'react-router-dom'
import { Loader2, TriangleAlert } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { useCourseStore, CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { getOwnCoursePhaseParticipation } from '@/network/queries/getOwnCoursePhaseParticipation'
import { useDevOpsChallengeStore } from './zustand/useDevOpsChallengeStore'
import { useGetDeveloperProfile } from './pages/hooks/useGetDeveloperProfile'
import UnauthorizedPage from '@/components/UnauthorizedPage'
import { Alert, AlertDescription, AlertTitle, ErrorPage } from '@tumaet/prompt-ui-components'

interface DevOpsChallengeDataShellProps {
  children: React.ReactNode
}

export const DevOpsChallengeDataShell = ({ children }: DevOpsChallengeDataShellProps) => {
  const { isStudentOfCourse } = useCourseStore()
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const { setCoursePhaseParticipation, setDeveloperProfile } = useDevOpsChallengeStore()

  const [developerProfileSet, setDeveloperProfileSet] = useState(false)
  const [participationSet, setParticipationSet] = useState(false)

  // getting the course phase participation
  const {
    data: fetchedParticipation,
    error: participationError,
    isPending: isParticipationPending,
    isError: isParticipationError,
    refetch: refetchParticipation,
  } = useQuery<CoursePhaseParticipationWithStudent>({
    queryKey: ['course_phase_participation', phaseId],
    queryFn: () => getOwnCoursePhaseParticipation(phaseId ?? ''),
  })

  // trying to get the developerProfile
  const {
    data: fetchedProfile,
    error: developerProfileError,
    isPending: isProfilePending,
    isError: isProfileError,
    refetch: refetchProfile,
  } = useGetDeveloperProfile()

  const isPending =
    isParticipationPending || isProfilePending || !developerProfileSet || !participationSet
  const isError = isParticipationError || isProfileError || !fetchedProfile

  useEffect(() => {
    if (fetchedParticipation) {
      setCoursePhaseParticipation(fetchedParticipation)
      setParticipationSet(true)
    }
  }, [fetchedParticipation, setCoursePhaseParticipation])

  useEffect(() => {
    if (
      !fetchedProfile ||
      (isProfileError && developerProfileError?.message.includes('student not found'))
    ) {
      setDeveloperProfile(undefined)
    } else if (fetchedProfile) {
      setDeveloperProfile(fetchedProfile)
    }
    setDeveloperProfileSet(true)
  }, [fetchedProfile, setDeveloperProfile, developerProfileError?.message, isProfileError])

  // if he is not a student -> we do not wait for the participation
  if (isStudent && isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  // Data only relevant for students - not for lecturers
  if (isStudent && isError) {
    // if the participation is not found, we show the unauthorized page bc then the student has not yet processed to this phase
    if (isParticipationError && participationError.message.includes('404')) {
      return <UnauthorizedPage backUrl={`/management/course/${courseId}`} />
    } else {
      if (!isProfileError) {
        return (
          <ErrorPage
            onRetry={() => {
              refetchProfile()
              refetchParticipation()
            }}
          />
        )
      }
    }
  }

  return (
    <>
      {!isStudent && (
        <Alert>
          <TriangleAlert className='h-4 w-4' />
          <AlertTitle>Your are not a student of this course.</AlertTitle>
          <AlertDescription>
            The following components are disabled because you are not a student of this course. For
            configuring this view, please refer to the Intro Course in the Tutor Course.
          </AlertDescription>
        </Alert>
      )}
      {children}
    </>
  )
}
