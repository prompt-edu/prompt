import { PassStatus } from '@tumaet/prompt-shared-state'
import { useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  parseSortParam,
  SORTING_QUERY_PARAM,
  type SortOption,
  serializeSortParam,
} from '../utils/sortOptions'
import { useParticipationStore } from '../zustand/useParticipationStore'

/** Keeps the overview sort in the URL so it survives reloads and the way back from a profile. */
export const useSortSearchParam = (): [SortOption, (sortBy: SortOption) => void] => {
  const [searchParams, setSearchParams] = useSearchParams()
  const sortBy = parseSortParam(searchParams.get(SORTING_QUERY_PARAM))

  const setSortBy = (nextSortBy: SortOption) => {
    setSearchParams(
      (params) => {
        const serialized = serializeSortParam(nextSortBy)
        if (serialized) {
          params.set(SORTING_QUERY_PARAM, serialized)
        } else {
          params.delete(SORTING_QUERY_PARAM)
        }
        return params
      },
      { replace: true },
    )
  }

  return [sortBy, setSortBy]
}

export const useSorting = (sortBy: SortOption) => {
  const { participations, interviewSlots, interviewReviews } = useParticipationStore()

  return useMemo(() => {
    return [...participations].sort((a, b) => {
      switch (sortBy) {
        case 'interviewDate': {
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
        }
        case 'firstName':
          return a.student.firstName.localeCompare(b.student.firstName)
        case 'lastName':
          return a.student.lastName.localeCompare(b.student.lastName)
        case 'acceptanceStatus': {
          const statusOrder = [PassStatus.PASSED, PassStatus.NOT_ASSESSED, PassStatus.FAILED]
          const aIndex = statusOrder.indexOf(a.passStatus)
          const bIndex = statusOrder.indexOf(b.passStatus)
          return (aIndex === -1 ? 999 : aIndex) - (bIndex === -1 ? 999 : bIndex)
        }
        case 'interviewScore':
          return (
            (interviewReviews[a.courseParticipationID]?.score || Number.MAX_VALUE) -
            (interviewReviews[b.courseParticipationID]?.score || Number.MAX_VALUE)
          )
        default:
          return 0
      }
    })
  }, [participations, sortBy, interviewSlots, interviewReviews])
}
