import { useQuery } from '@tanstack/react-query'
import { cn, ManagementPageHeader, useScreenSize } from '@tumaet/prompt-ui-components'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { SortDropdownMenu } from '../../components/SortDropdownMenu'
import { StudentCard } from '../../components/StudentCard'
import type { InterviewNavigationState } from '../../hooks/useInterviewNavigation'
import { useSorting, useSortSearchParam } from '../../hooks/useSorting'
import type { InterviewSlotWithAssignments } from '../../interfaces/InterviewSlots'
import { interviewAxiosInstance } from '../../network/interviewServerConfig'

export const OverviewPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const { pathname: path, search } = useLocation()
  const { width } = useScreenSize() // use this for more fine-grained control over the layout
  const [sortBy, setSortBy] = useSortSearchParam()
  const orderedParticipations = useSorting(sortBy)
  const navigationState: InterviewNavigationState = {
    orderedParticipationIds: orderedParticipations.map((p) => p.courseParticipationID),
  }

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
            onClick={() =>
              navigate(
                { pathname: `${path}/details/${participation.student.id}`, search },
                { state: navigationState },
              )
            }
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
