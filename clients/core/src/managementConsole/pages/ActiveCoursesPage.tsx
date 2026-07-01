import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useMemo, useState } from 'react'
import { CourseCards } from '../shared/components/CourseCard/CourseCards'
import { CourseTable } from '../shared/components/CourseTable/CourseTable'
import { CourseViewMode, CourseViewToggle } from '../shared/components/CourseViewToggle'

export const ActiveCoursesPage = () => {
  const { courses } = useCourseStore()
  const [viewMode, setViewMode] = useState<CourseViewMode>('cards')

  const activeCourses = useMemo(
    () => courses.filter((course) => !course.template && !course.archived),
    [courses],
  )

  const isEmpty = activeCourses.length == 0

  return (
    <div className='flex flex-col gap-6 w-full'>
      <div className='flex items-center justify-between gap-4 flex-wrap'>
        <h1 className='text-3xl font-bold tracking-tight'>Courses</h1>
        <CourseViewToggle viewMode={viewMode} onChange={setViewMode} />
      </div>

      {isEmpty ? (
        <p>No active courses yet</p>
      ) : viewMode === 'cards' ? (
        <CourseCards courses={activeCourses} />
      ) : (
        <CourseTable courses={activeCourses} />
      )}
    </div>
  )
}
