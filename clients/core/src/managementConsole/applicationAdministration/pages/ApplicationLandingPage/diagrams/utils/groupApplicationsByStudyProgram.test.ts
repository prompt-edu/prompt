import { PassStatus } from '@tumaet/prompt-shared-state'
import { describe, expect, it } from 'vitest'
import type { ApplicationParticipation } from '../../../../interfaces/applicationParticipation'
import { groupApplicationsByStudyProgram } from './groupApplicationsByStudyProgram'

const computerScience = { id: 'cs', name: 'Computer Science', shortName: 'CS' }
const physics = { id: 'ph', name: 'Physics', shortName: null }

const application = (
  studyProgram: string | undefined,
  passStatus: PassStatus = PassStatus.NOT_ASSESSED,
): ApplicationParticipation =>
  ({
    passStatus,
    student: { studyProgram },
  }) as ApplicationParticipation

const totals = (points: ReturnType<typeof groupApplicationsByStudyProgram>) =>
  points.map(({ program, dataKey, total }) => ({ program, dataKey, total }))

describe('groupApplicationsByStudyProgram', () => {
  it('gives every listed program a bar labeled with its short name or, failing that, its name', () => {
    const points = groupApplicationsByStudyProgram([], [computerScience, physics])

    expect(totals(points)).toEqual([
      { program: 'Computer Science', dataKey: 'CS', total: 0 },
      { program: 'Physics', dataKey: 'Physics', total: 0 },
      { program: 'Other', dataKey: 'Other', total: 0 },
    ])
  })

  it('counts a program that is not on the list as Other instead of dropping it', () => {
    const points = groupApplicationsByStudyProgram(
      [application('Robotics'), application('Computer Science'), application('Other')],
      [computerScience],
    )

    expect(totals(points)).toEqual([
      { program: 'Computer Science', dataKey: 'CS', total: 1 },
      { program: 'Other', dataKey: 'Other', total: 2 },
    ])
  })

  it('matches stored values after trimming them', () => {
    const points = groupApplicationsByStudyProgram(
      [application('  Computer Science ')],
      [computerScience],
    )

    expect(points[0].total).toBe(1)
  })

  it('adds an Unknown bar only when an application has no study program', () => {
    const points = groupApplicationsByStudyProgram(
      [application(''), application(undefined), application('   ')],
      [computerScience],
    )

    expect(totals(points)).toEqual([
      { program: 'Computer Science', dataKey: 'CS', total: 0 },
      { program: 'Other', dataKey: 'Other', total: 0 },
      { program: 'Unknown', dataKey: 'Unknown', total: 3 },
    ])
  })

  it('splits each bar by pass status', () => {
    const points = groupApplicationsByStudyProgram(
      [
        application('Computer Science', PassStatus.PASSED),
        application('Computer Science', PassStatus.PASSED),
        application('Computer Science', PassStatus.FAILED),
        application('Computer Science', PassStatus.NOT_ASSESSED),
      ],
      [computerScience],
    )

    expect(points[0]).toMatchObject({ accepted: 2, rejected: 1, notAssessed: 1, total: 4 })
  })
})
