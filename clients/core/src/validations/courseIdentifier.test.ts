import { describe, expect, it } from 'vitest'
import { copyCourseSchema } from './copyCourse'
import { courseFormSchema } from './course'
import { makeTemplateCourseSchema } from './makeTemplateCourse'
import { templateFormSchema } from './template'

const schemas = {
  courseFormSchema,
  copyCourseSchema,
  makeTemplateCourseSchema,
  templateFormSchema,
}

describe.each(Object.entries(schemas))('%s', (_, schema) => {
  it.each(['ioslab', 'iOS Lab 2025', 'Agile Project Management', 'Übung 日本語'])(
    'accepts the name %j',
    (name) => {
      expect(schema.shape.name.safeParse(name).success).toBe(true)
    },
  )

  it.each([
    '',
    'ios-lab',
    'ioslab/Lecturer',
    '100% Lab',
    'lab?1',
    'lab#1',
    'lab;1',
    'lab\t1',
    'lab\u007f',
  ])('rejects the name %j', (name) => {
    expect(schema.shape.name.safeParse(name).success).toBe(false)
  })

  it.each(['ws24', 'WS 2425'])('accepts the semester tag %j', (semesterTag) => {
    expect(schema.shape.semesterTag.safeParse(semesterTag).success).toBe(true)
  })

  it.each(['', 'ws-24', 'ws24/25', 'ws%24', 'ws?24', 'ws#24', 'ws;24', 'ws24\n'])(
    'rejects the semester tag %j',
    (semesterTag) => {
      expect(schema.shape.semesterTag.safeParse(semesterTag).success).toBe(false)
    },
  )
})
