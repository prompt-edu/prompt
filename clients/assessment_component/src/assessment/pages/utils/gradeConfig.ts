export const VALID_GRADE_VALUES = [1.0, 1.3, 1.7, 2.0, 2.3, 2.7, 3.0, 3.3, 3.7, 4.0, 5.0] as const

export type GradeValue = (typeof VALID_GRADE_VALUES)[number]

export const GRADE_CONFIG = {
  '1.0': { label: '1.0', color: '#60a5fa', longLabel: 'Very Good - 1.0' },
  '1.3': { label: '1.3', color: '#93c5fd', longLabel: 'Very Good - 1.3' },
  '1.7': { label: '1.7', color: '#4ade80', longLabel: 'Good - 1.7' },
  '2.0': { label: '2.0', color: '#86efac', longLabel: 'Good - 2.0' },
  '2.3': { label: '2.3', color: '#bbf7d0', longLabel: 'Good - 2.3' },
  '2.7': { label: '2.7', color: '#fef08a', longLabel: 'Satisfactory - 2.7' },
  '3.0': { label: '3.0', color: '#fde68a', longLabel: 'Satisfactory - 3.0' },
  '3.3': { label: '3.3', color: '#fcd34d', longLabel: 'Satisfactory - 3.3' },
  '3.7': { label: '3.7', color: '#fb923c', longLabel: 'Sufficient - 3.7' },
  '4.0': { label: '4.0', color: '#f97316', longLabel: 'Sufficient - 4.0' },
  '5.0': { label: '5.0', color: '#fca5a5', longLabel: 'Fail - 5.0' },
  'No Grade': { label: 'No Grade', color: '#d4d4d8', longLabel: 'No Grade' },
} as const

export const GRADE_SELECT_OPTIONS = VALID_GRADE_VALUES.map((value) => ({
  value: value.toFixed(1),
  label: GRADE_CONFIG[value.toFixed(1) as keyof typeof GRADE_CONFIG].longLabel,
}))

export const getGradeColor = (grade: number): string => {
  const gradeKey =
    VALID_GRADE_VALUES.find((validGrade) => grade <= validGrade)?.toFixed(1) || 'No Grade'
  return GRADE_CONFIG[gradeKey]?.color || '#d4d4d8'
}

export const validateGrade = (
  gradeString: string,
): { isValid: boolean; value?: number; error?: string } => {
  if (!gradeString || gradeString.trim() === '') {
    return { isValid: true, value: 5.0 }
  }

  const gradeValue = parseFloat(gradeString)

  if (Number.isNaN(gradeValue)) {
    return { isValid: false, error: 'Grade must be a valid number' }
  }

  if (gradeValue < 1 || gradeValue > 5) {
    return { isValid: false, error: 'Grade must be between 1.0 and 5.0' }
  }

  const isValidValue = VALID_GRADE_VALUES.some(
    (validValue) => Math.abs(validValue - gradeValue) < 0.01,
  )

  if (!isValidValue) {
    return {
      isValid: false,
      error: `Grade must be one of the predefined values (${VALID_GRADE_VALUES.join(', ')})`,
    }
  }

  return { isValid: true, value: gradeValue }
}

export const isValidGrade = (grade: number): grade is GradeValue => {
  const validation = validateGrade(grade.toString())
  return validation.isValid
}

export const GRADE_RANGE = {
  MIN: 1.0,
  MAX: 5.0,
  TOTAL_RANGE: 4.0,
} as const
