import type { StudyProgram } from '@core/managementConsole/shared/interfaces/StudyProgram'
import {
  OTHER_STUDY_PROGRAM,
  UNKNOWN_STUDY_PROGRAM,
} from '@core/managementConsole/shared/utils/otherStudyProgram'
import { PassStatus } from '@tumaet/prompt-shared-state'
import type { ApplicationParticipation } from '../../../../interfaces/applicationParticipation'
import type { DataPoint } from '../../interfaces/DataPoint'

export interface StudyProgramDataPoint extends DataPoint {
  program: string
  accepted: number
  rejected: number
  notAssessed: number
}

type PassStatusCounts = Record<PassStatus, number>

const emptyCounts = (): PassStatusCounts => ({
  [PassStatus.PASSED]: 0,
  [PassStatus.FAILED]: 0,
  [PassStatus.NOT_ASSESSED]: 0,
})

const dataPoint = (
  program: string,
  label: string,
  counts: PassStatusCounts,
): StudyProgramDataPoint => {
  const accepted = counts[PassStatus.PASSED]
  const rejected = counts[PassStatus.FAILED]
  const notAssessed = counts[PassStatus.NOT_ASSESSED]
  return {
    program,
    dataKey: label,
    accepted,
    rejected,
    notAssessed,
    total: accepted + rejected + notAssessed,
  }
}

export const groupApplicationsByStudyProgram = (
  applications: ApplicationParticipation[],
  studyPrograms: StudyProgram[],
): StudyProgramDataPoint[] => {
  const listedCounts = new Map(
    studyPrograms.map((studyProgram) => [studyProgram.name, emptyCounts()]),
  )
  const otherCounts = emptyCounts()
  const unknownCounts = emptyCounts()
  let hasUnknown = false

  for (const application of applications) {
    const studyProgram = application.student.studyProgram?.trim() ?? ''
    if (studyProgram === '') {
      hasUnknown = true
      unknownCounts[application.passStatus] += 1
    } else {
      const counts = listedCounts.get(studyProgram) ?? otherCounts
      counts[application.passStatus] += 1
    }
  }

  const listed = studyPrograms.map((studyProgram) =>
    dataPoint(
      studyProgram.name,
      studyProgram.shortName ?? studyProgram.name,
      listedCounts.get(studyProgram.name) ?? emptyCounts(),
    ),
  )
  const other = dataPoint(OTHER_STUDY_PROGRAM, OTHER_STUDY_PROGRAM, otherCounts)
  const unknown = hasUnknown
    ? [dataPoint(UNKNOWN_STUDY_PROGRAM, UNKNOWN_STUDY_PROGRAM, unknownCounts)]
    : []

  return [...listed, other, ...unknown]
}
