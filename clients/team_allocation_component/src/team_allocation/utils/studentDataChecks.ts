import type { TeaseStudent } from '../interfaces/tease/student'
import type { ValidationResult } from '../interfaces/validationResult'
import { checksConfig } from './checksConfig'
import { hasSubmittedSurvey } from './surveyStatus'

export interface LaunchWarnings {
  incompleteChecks: string[]
  studentsWithoutSurvey: number
  totalStudents: number
}

export const runStudentDataChecks = (students: TeaseStudent[]): ValidationResult[] =>
  checksConfig.map(
    ({
      label,
      extractor,
      isEmpty,
      missingMessage,
      problemDescription,
      details,
      category,
      icon,
      highLevelCategory,
    }) => {
      const validStudents = students.filter((s) => !isEmpty(extractor(s)))
      const completionRate =
        students.length > 0 ? Math.round((validStudents.length / students.length) * 100) : 0
      const allValid = validStudents.length === students.length && students.length > 0
      const noneValid = validStudents.length === 0

      return {
        label,
        isValid: allValid,
        category,
        highLevelCategory,
        completionRate,
        icon,
        details: allValid || noneValid ? details : undefined,
        problemDescription: allValid
          ? undefined
          : noneValid
            ? problemDescription
            : `${validStudents.length} of ${students.length} students have provided ${missingMessage}.`,
      }
    },
  )

export const getLaunchWarnings = (students: TeaseStudent[]): LaunchWarnings => ({
  incompleteChecks:
    students.length > 0
      ? runStudentDataChecks(students)
          .filter((check) => !check.isValid)
          .map((check) => check.label)
      : [],
  studentsWithoutSurvey: students.filter((s) => !hasSubmittedSurvey(s)).length,
  totalStudents: students.length,
})
