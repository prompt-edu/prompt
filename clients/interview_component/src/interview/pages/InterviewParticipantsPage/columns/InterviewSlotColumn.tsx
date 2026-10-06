import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import type {
  ExtraParticipantColumn,
  ParticipantRow,
  TableFilter,
} from '@tumaet/prompt-ui-components'
import { format } from 'date-fns'
import type { InterviewSlot } from '../../../interfaces/InterviewSlots'
import {
  compareNullableNumbers,
  getInterviewDayLabel,
  getInterviewDayOptions,
  matchesSelect,
  NOT_SCHEDULED,
} from '../utils/tableFns'

const COLUMN_ID = 'interviewSlot'

const getScheduledSlot = (slot: InterviewSlot | undefined) =>
  slot?.startTime && slot.endTime
    ? { start: new Date(slot.startTime), end: new Date(slot.endTime) }
    : undefined

const getStartTime = (slot: InterviewSlot | undefined): number | null =>
  getScheduledSlot(slot)?.start.getTime() ?? null

export const createInterviewSlotColumn = (
  participations: CoursePhaseParticipationWithStudent[],
  slotByParticipation: Map<string, InterviewSlot>,
): ExtraParticipantColumn<number | null> => {
  const getParticipantStartTime = (courseParticipationID: string) =>
    getStartTime(slotByParticipation.get(courseParticipationID))

  const formatSlot = (courseParticipationID: string): string => {
    const slot = getScheduledSlot(slotByParticipation.get(courseParticipationID))
    return slot
      ? `${format(slot.start, 'PPP')}, ${format(slot.start, 'p')} - ${format(slot.end, 'p')}`
      : NOT_SCHEDULED
  }

  return {
    id: COLUMN_ID,
    header: 'Interview',

    accessorFn: (row: ParticipantRow) => getParticipantStartTime(row.courseParticipationID),

    cell: ({ row }) => {
      const slot = getScheduledSlot(slotByParticipation.get(row.original.courseParticipationID))
      if (!slot) return <span className='text-muted-foreground'>{NOT_SCHEDULED}</span>
      return (
        <div className='flex flex-col whitespace-nowrap'>
          <span>{format(slot.start, 'PPP')}</span>
          <span className='text-xs text-muted-foreground'>
            {format(slot.start, 'p')} - {format(slot.end, 'p')}
          </span>
        </div>
      )
    },

    enableSorting: true,
    sortFn: (rowA, rowB) =>
      compareNullableNumbers(
        getParticipantStartTime(rowA.original.courseParticipationID),
        getParticipantStartTime(rowB.original.courseParticipationID),
      ),

    enableColumnFilter: true,
    filterFn: (row, _columnId, filterValue) =>
      matchesSelect(
        getInterviewDayLabel(getParticipantStartTime(row.original.courseParticipationID)),
        filterValue,
      ),

    extraData: participations.map((p) => ({
      courseParticipationID: p.courseParticipationID,
      value: getParticipantStartTime(p.courseParticipationID),
      stringValue: formatSlot(p.courseParticipationID),
    })),
  }
}

export const createInterviewDayFilter = (
  participations: CoursePhaseParticipationWithStudent[],
  slotByParticipation: Map<string, InterviewSlot>,
): TableFilter<ParticipantRow> => ({
  type: 'select',
  id: COLUMN_ID,
  label: 'Interview Day',
  options: getInterviewDayOptions(
    participations.map((p) => getStartTime(slotByParticipation.get(p.courseParticipationID))),
  ),
})
