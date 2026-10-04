import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import type { ExtraParticipantColumn, ParticipantRow } from '@tumaet/prompt-ui-components'
import { ExternalLink } from 'lucide-react'
import { Link } from 'react-router-dom'

export const createApplicationLinkColumn = (
  participations: CoursePhaseParticipationWithStudent[],
  getApplicationPath: (courseParticipationID: string) => string | undefined,
): ExtraParticipantColumn<string | null> | undefined => {
  if (!participations.some((p) => getApplicationPath(p.courseParticipationID))) return undefined

  return {
    id: 'application',
    header: 'Application',

    accessorFn: (row: ParticipantRow) => getApplicationPath(row.courseParticipationID) ?? null,

    cell: ({ getValue }) => {
      const path = getValue()
      if (!path) return null
      return (
        <Link
          to={path}
          onClick={(e) => e.stopPropagation()}
          className='inline-flex items-center gap-1 whitespace-nowrap text-primary underline-offset-4 hover:underline'
        >
          <ExternalLink className='h-3 w-3' />
          View application
        </Link>
      )
    },

    enableSorting: false,

    extraData: participations.map((p) => {
      const path = getApplicationPath(p.courseParticipationID)
      return {
        courseParticipationID: p.courseParticipationID,
        value: path ?? null,
        stringValue: path ? `${window.location.origin}${path}` : '',
      }
    }),
  }
}
