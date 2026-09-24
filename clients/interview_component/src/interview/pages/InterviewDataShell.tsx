import { useQuery } from '@tanstack/react-query'
import {
  type CoursePhaseParticipationsWithResolution,
  type CoursePhaseWithMetaData,
  getCoursePhase,
  getCoursePhaseParticipations,
} from '@tumaet/prompt-shared-state'
import { QueryGate } from '@tumaet/prompt-ui-components'
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
  const coursePhaseParticipationsQuery = useQuery<CoursePhaseParticipationsWithResolution>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
  })

  const coursePhaseQuery = useQuery<CoursePhaseWithMetaData>({
    queryKey: ['course_phase', phaseId],
    queryFn: () => getCoursePhase(phaseId ?? ''),
  })

  // Fetch interview slots with assignments from the interview server
  const interviewSlotsQuery = useQuery<InterviewSlotWithAssignments[]>({
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
  const interviewReviewsQuery = useQuery<InterviewReview[]>({
    queryKey: ['interviewReviews', phaseId],
    queryFn: () => getInterviewReviews(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const { data: coursePhaseParticipations } = coursePhaseParticipationsQuery
  const { data: coursePhase } = coursePhaseQuery
  const { data: interviewSlotsWithAssignments } = interviewSlotsQuery
  const { data: interviewReviews } = interviewReviewsQuery

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
    <QueryGate
      queries={[
        coursePhaseParticipationsQuery,
        coursePhaseQuery,
        interviewSlotsQuery,
        interviewReviewsQuery,
      ]}
    >
      {children}
    </QueryGate>
  )
}
