import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import type { ExtraParticipantColumn } from '@tumaet/prompt-ui-components'
import { Check, X } from 'lucide-react'
import type { TeaseStudent } from '../../../interfaces/tease/student'
import {
  getSurveyStatus,
  SURVEY_STATUS_COLUMN_ID,
  SURVEY_SUBMITTED,
} from '../../../utils/surveyStatus'

export const createSurveyStatusColumn = (
  participations: CoursePhaseParticipationWithStudent[],
  teaseStudentByParticipation: Map<string, TeaseStudent>,
): ExtraParticipantColumn<string> => ({
  id: SURVEY_STATUS_COLUMN_ID,
  header: 'Survey',

  accessorFn: (row) => getSurveyStatus(teaseStudentByParticipation.get(row.courseParticipationID)),

  cell: ({ getValue }) => {
    const status = getValue()
    const submitted = status === SURVEY_SUBMITTED

    return (
      <div className='flex items-center gap-2'>
        <div
          className={
            submitted
              ? 'bg-green-100 text-green-700 p-1 rounded-full'
              : 'bg-red-100 text-red-700 p-1 rounded-full'
          }
        >
          {submitted ? <Check className='h-4 w-4' /> : <X className='h-4 w-4' />}
        </div>
        <span>{status}</span>
      </div>
    )
  },

  enableSorting: true,

  enableColumnFilter: true,
  filterFn: (row, columnId, filterValue) => {
    if (!Array.isArray(filterValue) || filterValue.length === 0) return true
    return filterValue.includes(String(row.getValue(columnId)))
  },

  extraData: participations.map((participation) => {
    const status = getSurveyStatus(
      teaseStudentByParticipation.get(participation.courseParticipationID),
    )

    return {
      courseParticipationID: participation.courseParticipationID,
      value: status,
      stringValue: status,
    }
  }),
})
