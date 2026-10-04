import type {
  StudyProgram,
  StudyProgramStudentCount,
} from '../../../shared/interfaces/StudyProgram'
import { OTHER_STUDY_PROGRAM, UNKNOWN_STUDY_PROGRAM } from '../../../shared/utils/otherStudyProgram'

export interface StudyProgramWithStudentCount extends StudyProgram {
  studentCount: number | null
}

const RESERVED_STUDY_PROGRAMS = [
  { name: OTHER_STUDY_PROGRAM, reason: 'study programs applicants enter as free text' },
  { name: UNKNOWN_STUDY_PROGRAM, reason: 'applications without a study program' },
]

export const reservedStudyProgramError = (value: string): string | undefined => {
  const normalized = value.trim().toLowerCase()
  const reserved = RESERVED_STUDY_PROGRAMS.find(({ name }) => name.toLowerCase() === normalized)
  return reserved && `"${reserved.name}" is reserved for ${reserved.reason}.`
}

export const withStudentCounts = (
  studyPrograms: StudyProgram[],
  counts: StudyProgramStudentCount[] | undefined,
): StudyProgramWithStudentCount[] => {
  const countById = new Map(counts?.map((count) => [count.studyProgramID, count.studentCount]))
  return studyPrograms.map((studyProgram) => ({
    ...studyProgram,
    studentCount: counts === undefined ? null : (countById.get(studyProgram.id) ?? 0),
  }))
}

const students = (count: number): string => `${count} student${count === 1 ? '' : 's'}`

export const renameWarning = (studentCount: number, previousName: string, newName: string) =>
  `${students(studentCount)} with "${previousName}" will be updated to "${newName}".`

export const deleteWarning = (studyPrograms: StudyProgramWithStudentCount[]): string => {
  const keepTheirProgram = `keep their study program, which will then count as "${OTHER_STUDY_PROGRAM}".`
  if (studyPrograms.some((studyProgram) => studyProgram.studentCount === null)) {
    return `Student counts could not be loaded. Students ${keepTheirProgram}`
  }
  const studentCount = studyPrograms.reduce(
    (sum, studyProgram) => sum + (studyProgram.studentCount ?? 0),
    0,
  )
  if (studentCount === 0) {
    return 'No student has this study program.'
  }
  return `${students(studentCount)} ${keepTheirProgram}`
}
