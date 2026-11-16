import { type CoursePhaseWithType, useCourseStore } from '@tumaet/prompt-shared-state'
import { useMemo } from 'react'

const sortKey = (sequenceOrder: number): number =>
  sequenceOrder === -1 ? Number.MAX_SAFE_INTEGER : sequenceOrder

export const useCoursePhases = (courseId: string | undefined): CoursePhaseWithType[] => {
  const { courses } = useCourseStore()
  return useMemo(
    () =>
      [...(courses.find((course) => course.id === courseId)?.coursePhases ?? [])].sort(
        (a, b) => sortKey(a.sequenceOrder) - sortKey(b.sequenceOrder),
      ),
    [courseId, courses],
  )
}
