import { useQuery } from '@tanstack/react-query'
import {
  type CoursePhaseParticipationsWithResolution,
  type CoursePhaseWithMetaData,
  getCoursePhase,
  getCoursePhaseParticipations,
} from '@tumaet/prompt-shared-state'
import { ErrorPage } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import type { InterviewReview } from '../interfaces/InterviewReview'
import type { InterviewSlot, InterviewSlotWithAssignments } from '../interfaces/InterviewSlots'
import { interviewAxiosInstance } from '../network/interviewServerConfig'
import { getInterviewReviews } from '../network/queries/getInterviewReviews'
import { useCoursePhaseStore } from '../zustand/useCoursePhaseStore'
import { useParticipationStore } from '../zustand/useParticipationStore'

interface InterviewDataShellProps {
  children: React.ReactNode
}

export const InterviewDataShell = ({ children }: InterviewDataShellProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { setParticipations, setInterviewSlots, setInterviewReviews } = useParticipationStore()
  const { setCoursePhase } = useCoursePhaseStore()
  const {
    data: coursePhaseParticipations,
    isPending: isCoursePhaseParticipationsPending,
    isError: isParticipationsError,
    refetch: refetchCoursePhaseParticipations,
  } = useQuery<CoursePhaseParticipationsWithResolution>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
  })

  const {
    data: coursePhase,
    isPending: isCoursePhasePending,
    isError: isCoursePhaseError,
    refetch: refetchCoursePhase,
  } = useQuery<CoursePhaseWithMetaData>({
    queryKey: ['course_phase', phaseId],
    queryFn: () => getCoursePhase(phaseId ?? ''),
  })

  // Fetch interview slots with assignments from the interview server
  const {
    data: interviewSlotsWithAssignments,
    isPending: isInterviewSlotsPending,
    isError: isInterviewSlotsError,
    refetch: refetchInterviewSlots,
  } = useQuery<InterviewSlotWithAssignments[]>({
    queryKey: ['interviewSlotsWithAssignments', phaseId],
    queryFn: async () => {
      const response = await interviewAxiosInstance.get(
        `interview/api/course_phase/${phaseId}/interview-slots`,
      )
      return response.data
    },
    enabled: !!phaseId,
  })

  // Fetch interview reviews (score, interviewer, answers) from the interview server
  const {
    data: interviewReviews,
    isPending: isInterviewReviewsPending,
    isError: isInterviewReviewsError,
    refetch: refetchInterviewReviews,
  } = useQuery<InterviewReview[]>({
    queryKey: ['interviewReviews', phaseId],
    queryFn: () => getInterviewReviews(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const isError =
    isParticipationsError || isCoursePhaseError || isInterviewSlotsError || isInterviewReviewsError
  const isPending =
    isCoursePhaseParticipationsPending ||
    isCoursePhasePending ||
    isInterviewSlotsPending ||
    isInterviewReviewsPending
  const refetch = () => {
    refetchCoursePhaseParticipations()
    refetchCoursePhase()
    refetchInterviewSlots()
    refetchInterviewReviews()
  }

  useEffect(() => {
    if (coursePhaseParticipations) {
      setParticipations(coursePhaseParticipations.participations)
    }
  }, [coursePhaseParticipations, setParticipations])

  useEffect(() => {
    if (coursePhase) {
      setCoursePhase(coursePhase)
    }
  }, [coursePhase, setCoursePhase])

  useEffect(() => {
    if (interviewSlotsWithAssignments) {
      // Transform the slots with assignments into a flat array
      // where each assignment creates an entry with courseParticipationID
      const flattenedSlots: InterviewSlot[] = []

      interviewSlotsWithAssignments.forEach((slot) => {
        slot.assignments.forEach((assignment) => {
          flattenedSlots.push({
            id: slot.id,
            startTime: slot.startTime,
            endTime: slot.endTime,
            courseParticipationID: assignment.courseParticipationId,
          })
        })
      })

      setInterviewSlots(flattenedSlots)
    }
  }, [interviewSlotsWithAssignments, setInterviewSlots])

  useEffect(() => {
    if (interviewReviews) {
      setInterviewReviews(interviewReviews)
    }
  }, [interviewReviews, setInterviewReviews])

  return (
    <>
      {isError ? (
        <ErrorPage onRetry={refetch} />
      ) : isPending ? (
        <div className='flex justify-center items-center h-64'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
        </div>
      ) : (
        children
      )}
    </>
  )
}
