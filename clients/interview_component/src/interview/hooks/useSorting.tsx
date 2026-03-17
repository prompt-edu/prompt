import { useMemo } from 'react'
import { useParticipationStore } from '../zustand/useParticipationStore'
import { PassStatus } from '@tumaet/prompt-shared-state'

export const useSorting = (sortBy: string | undefined) => {
  const { participations, interviewSlots } = useParticipationStore()

  return useMemo(() => {
    if (!sortBy) return participations
    return [...participations].sort((a, b) => {
      switch (sortBy) {
        case 'Interview Date':
          const aSlot = interviewSlots.find(
            (slot) => slot.courseParticipationID === a.courseParticipationID,
          )
          const bSlot = interviewSlots.find(
            (slot) => slot.courseParticipationID === b.courseParticipationID,
          )
          // Sort by startTime (ascending - earlier slots first)
          // Participations without slots go to the end
          const aStartTime = aSlot?.startTime
          const bStartTime = bSlot?.startTime

          if (!aStartTime && !bStartTime) return 0
          if (!aStartTime) return 1
          if (!bStartTime) return -1
          const timeComparison = new Date(aStartTime).getTime() - new Date(bStartTime).getTime()
          // If times are equal, sort by last name, then first name for consistency
          if (timeComparison === 0) {
            const lastNameComparison = a.student.lastName.localeCompare(b.student.lastName)
            if (lastNameComparison !== 0) return lastNameComparison
            return a.student.firstName.localeCompare(b.student.firstName)
          }
          return timeComparison
        case 'First Name':
          return a.student.firstName.localeCompare(b.student.firstName)
        case 'Last Name':
          return a.student.lastName.localeCompare(b.student.lastName)
        case 'Acceptance Status':
          const statusOrder = [PassStatus.PASSED, PassStatus.NOT_ASSESSED, PassStatus.FAILED]
          const aIndex = statusOrder.indexOf(a.passStatus)
          const bIndex = statusOrder.indexOf(b.passStatus)
          return (aIndex === -1 ? 999 : aIndex) - (bIndex === -1 ? 999 : bIndex)
        case 'Interview Score':
          return (
            (a.restrictedData.score || Number.MAX_VALUE) -
            (b.restrictedData.score || Number.MAX_VALUE)
          )
        default:
          return 0
      }
    })
  }, [participations, sortBy, interviewSlots])
}
