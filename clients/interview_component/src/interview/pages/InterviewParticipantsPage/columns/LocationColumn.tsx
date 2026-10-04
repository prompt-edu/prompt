import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import type { ExtraParticipantColumn, ParticipantRow } from '@tumaet/prompt-ui-components'
import type { InterviewSlot } from '../../../interfaces/InterviewSlots'
import { compareNullableStrings, isUrl } from '../utils/tableFns'

export const createLocationColumn = (
  participations: CoursePhaseParticipationWithStudent[],
  slotByParticipation: Map<string, InterviewSlot>,
): ExtraParticipantColumn<string | null> => {
  const getLocation = (courseParticipationID: string): string | null =>
    slotByParticipation.get(courseParticipationID)?.location || null

  return {
    id: 'interviewLocation',
    header: 'Location',

    accessorFn: (row: ParticipantRow) => getLocation(row.courseParticipationID),

    cell: ({ getValue }) => {
      const location = getValue()
      if (!location) return null
      if (!isUrl(location)) return location
      return (
        <a
          href={location}
          target='_blank'
          rel='noopener noreferrer'
          onClick={(e) => e.stopPropagation()}
          className='text-primary underline-offset-4 hover:underline'
        >
          {location}
        </a>
      )
    },

    enableSorting: true,
    sortFn: (rowA, rowB) =>
      compareNullableStrings(
        getLocation(rowA.original.courseParticipationID),
        getLocation(rowB.original.courseParticipationID),
      ),

    extraData: participations.map((p) => {
      const location = getLocation(p.courseParticipationID)
      return {
        courseParticipationID: p.courseParticipationID,
        value: location,
        stringValue: location ?? '',
      }
    }),
  }
}
