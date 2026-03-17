import { ColumnDef } from '@tanstack/react-table'
import type { AccessorFn, SortingFn, FilterFn } from '@tanstack/react-table'
import {
  CoursePhaseParticipationWithStudent,
  PassStatus,
  Student,
} from '@tumaet/prompt-shared-state'

export interface ParticipantRow {
  id: string

  coursePhaseID: string
  courseParticipationID: string
  passStatus: PassStatus

  restrictedData: Record<string, any>
  studentReadableData?: Record<string, any>
  prevData?: Record<string, any>

  student: Student

  firstName: string
  lastName: string
  email?: string
  matriculationNumber?: string
  universityLogin?: string

  [key: string]: unknown
}

export interface ExtraParticipantColumn<TValue> {
  id: string
  header: string

  accessorFn?: AccessorFn<ParticipantRow, TValue>
  cell?: ColumnDef<ParticipantRow, TValue>['cell']

  enableSorting?: boolean
  sortingFn?: SortingFn<ParticipantRow>

  enableColumnFilter?: boolean
  filterFn?: FilterFn<ParticipantRow>

  extraData: {
    courseParticipationID: string
    value: TValue
    stringValue?: string
  }[]
}

export function buildParticipantRows(
  participants: CoursePhaseParticipationWithStudent[],
  extraColumns: ExtraParticipantColumn<any>[],
): ParticipantRow[] {
  return participants.map((p) => {
    const extraValues: Record<string, unknown> = {}

    for (const col of extraColumns) {
      extraValues[col.id] =
        col.extraData.find((d) => d.courseParticipationID === p.courseParticipationID)?.value ??
        null
    }

    return {
      id: p.courseParticipationID,

      coursePhaseID: p.coursePhaseID,
      courseParticipationID: p.courseParticipationID,
      passStatus: p.passStatus,

      restrictedData: p.restrictedData,
      studentReadableData: p.studentReadableData,
      prevData: p.prevData,

      student: p.student,

      firstName: p.student.firstName,
      lastName: p.student.lastName,
      email: p.student.email,
      matriculationNumber: p.student.matriculationNumber,
      universityLogin: p.student.universityLogin,

      ...extraValues,
    }
  })
}
