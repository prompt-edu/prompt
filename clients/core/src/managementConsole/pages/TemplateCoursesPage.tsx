import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useMemo, useState } from 'react'
import { CourseCards } from '../shared/components/CourseCard/CourseCards'
import { CourseTable } from '../shared/components/CourseTable/CourseTable'
import { CourseViewMode, CourseViewToggle } from '../shared/components/CourseViewToggle'

export const TemplateCoursesPage = () => {
  const { courses } = useCourseStore()
  const [viewMode, setViewMode] = useState<CourseViewMode>('cards')

  const filteredCourses = useMemo(() => courses.filter((course) => course.template), [courses])

  return (
    <div className='flex flex-col gap-6 w-full'>
      <div className='flex items-center justify-between gap-4 flex-wrap'>
        <h1 className='text-3xl font-bold tracking-tight'>Template Courses</h1>
        <CourseViewToggle viewMode={viewMode} onChange={setViewMode} />
      </div>

      {viewMode === 'cards' ? (
        <>
          {filteredCourses.length == 0 && <p>No template courses yet</p>}
          <CourseCards courses={filteredCourses} />
        </>
      ) : (
        <CourseTable courses={filteredCourses} />
      )}
    </div>
  )
}
