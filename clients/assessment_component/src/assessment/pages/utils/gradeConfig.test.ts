import { describe, expect, it } from 'vitest'

import { isValidGrade, validateGrade } from './gradeConfig'

describe('validateGrade', () => {
  it('requires a grade', () => {
    expect(validateGrade('')).toEqual({ isValid: false, error: 'a grade must be selected' })
    expect(validateGrade('   ')).toEqual({ isValid: false, error: 'a grade must be selected' })
  })

  it('rejects a value that is not a number', () => {
    expect(validateGrade('very good')).toEqual({
      isValid: false,
      error: 'Grade must be a valid number',
    })
  })

  it('rejects a value outside the German grade range', () => {
    expect(validateGrade('0.9').error).toBe('Grade must be between 1.0 and 5.0')
    expect(validateGrade('5.1').error).toBe('Grade must be between 1.0 and 5.0')
  })

  it('rejects a number inside the range that is not a grade', () => {
    expect(validateGrade('4.3').error).toMatch(/predefined values/)
    expect(validateGrade('4.5').error).toMatch(/predefined values/)
  })

  it('accepts every predefined grade', () => {
    expect(validateGrade('1.0')).toEqual({ isValid: true, value: 1.0 })
    expect(validateGrade('1.7')).toEqual({ isValid: true, value: 1.7 })
    expect(validateGrade('4.0')).toEqual({ isValid: true, value: 4.0 })
    expect(validateGrade('5.0')).toEqual({ isValid: true, value: 5.0 })
  })

  it('accepts a grade written without a trailing decimal', () => {
    expect(validateGrade('2')).toEqual({ isValid: true, value: 2 })
  })
})

describe('isValidGrade', () => {
  it('narrows a number to a grade value', () => {
    expect(isValidGrade(3.3)).toBe(true)
    expect(isValidGrade(3.5)).toBe(false)
  })
})
