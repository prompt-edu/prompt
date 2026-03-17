import { useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { cn, ManagementPageHeader, useScreenSize } from '@tumaet/prompt-ui-components'
import { StudentCard } from '../../components/StudentCard'
import { SortDropdownMenu } from '../../components/SortDropdownMenu'
import { useSorting } from '../../hooks/useSorting'
import { useQuery } from '@tanstack/react-query'
import { interviewAxiosInstance } from '../../network/interviewServerConfig'
import { InterviewSlotWithAssignments } from '../../interfaces/InterviewSlots'

export const OverviewPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const path = useLocation().pathname
  const { width } = useScreenSize() // use this for more fine-grained control over the layout
  const [sortBy, setSortBy] = useState<string | undefined>('Interview Date')
  const orderedParticipations = useSorting(sortBy)

  // Fetch interview slots with assignments
  const { data: slots } = useQuery<InterviewSlotWithAssignments[]>({
    queryKey: ['interviewSlotsWithAssignments', phaseId],
    queryFn: async () => {
      const response = await interviewAxiosInstance.get(
        `interview/api/course_phase/${phaseId}/interview-slots`,
      )
      return response.data
    },
    enabled: !!phaseId,
  })

  // Create a map of participation ID to interview slot
  const participationToSlot = new Map<string, InterviewSlotWithAssignments>()
  slots?.forEach((slot) => {
    slot.assignments.forEach((assignment) => {
      participationToSlot.set(assignment.courseParticipationId, slot)
    })
  })

  return (
    <div>
      <ManagementPageHeader>Interview</ManagementPageHeader>
      <div className='flex justify-between items-center mt-4 mb-6'>
        <div className='flex space-x-2'>
          <SortDropdownMenu sortBy={sortBy} setSortBy={setSortBy} />
        </div>
      </div>
      <div
        className={cn(
          'grid gap-2 mt-8',
          width >= 1500 ? 'grid-cols-4' : width >= 1200 ? 'grid-cols-3' : 'grid-cols-2',
        )}
      >
        {orderedParticipations?.map((participation) => (
          <div
            key={participation.student.email}
            onClick={() => navigate(`${path}/details/${participation.student.id}`)}
            className='cursor-pointer'
          >
            <StudentCard
              participation={participation}
              interviewSlot={participationToSlot.get(participation.courseParticipationID)}
            />
          </div>
        ))}
      </div>
    </div>
  )
}

export default OverviewPage
