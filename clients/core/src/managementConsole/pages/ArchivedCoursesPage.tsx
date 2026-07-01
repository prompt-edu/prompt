import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useMemo, useState } from 'react'
import { CourseCards } from '../shared/components/CourseCard/CourseCards'
import { CourseTable } from '../shared/components/CourseTable/CourseTable'
import { type CourseViewMode, CourseViewToggle } from '../shared/components/CourseViewToggle'

export const ArchivedCoursesPage = () => {
  const { courses } = useCourseStore()
  const [viewMode, setViewMode] = useState<CourseViewMode>('table')

  const filteredCourses = useMemo(() => courses.filter((course) => course.archived), [courses])

  const isEmpty = filteredCourses.length === 0

  return (
    <div className='flex flex-col gap-6 w-full'>
      <div className='flex items-center justify-between gap-4 flex-wrap'>
        <h1 className='text-3xl font-bold tracking-tight'>Archived Courses</h1>
        <CourseViewToggle viewMode={viewMode} onChange={setViewMode} />
      </div>

      {isEmpty ? (
        <p>No Archived courses yet</p>
      ) : viewMode === 'cards' ? (
        <CourseCards courses={filteredCourses} />
      ) : (
        <CourseTable courses={filteredCourses} />
      )}
    </div>
  )
}
