import { VALID_GRADE_VALUES } from '../../../../utils/gradeConfig'

import { computeQuartile } from '../../utils/computeQuartile'

import type { GradeDistributionDataPoint } from '../interfaces/GradeDistributionDataPoint'

export const createGradeDistributionDataPoint = (
  shortLabel: string,
  label: string,
  grades: number[],
): GradeDistributionDataPoint => {
  if (grades.length === 0) {
    const emptyCounts = VALID_GRADE_VALUES.reduce(
      (acc, grade) => {
        acc[grade.toFixed(1)] = 0
        return acc
      },
      {} as Record<string, number>,
    )

    return {
      shortLabel,
      label,
      average: 0,
      lowerQuartile: 0,
      median: 0,
      upperQuartile: 0,
      counts: emptyCounts,
    }
  }

  const gradeCounts = VALID_GRADE_VALUES.reduce(
    (acc, gradeValue) => {
      const count = grades.filter((grade) => Math.abs(grade - gradeValue) < 0.01).length
      acc[gradeValue.toFixed(1)] = count
      return acc
    },
    {} as Record<string, number>,
  )

  return {
    shortLabel,
    label,
    average: grades.reduce((sum, grade) => sum + grade, 0) / grades.length,
    lowerQuartile: computeQuartile(grades, 0.25),
    median: computeQuartile(grades, 0.5),
    upperQuartile: computeQuartile(grades, 0.75),
    counts: gradeCounts,
  }
}
