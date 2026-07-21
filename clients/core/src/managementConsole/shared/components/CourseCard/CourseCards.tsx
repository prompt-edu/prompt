import type { Course } from '@tumaet/prompt-shared-state'
import { Input } from '@tumaet/prompt-ui-components'
import { AnimatePresence, motion } from 'framer-motion'
import { Search } from 'lucide-react'
import { useMemo, useState } from 'react'
import { CourseCard } from './CourseCard'

interface CourseCardsProps {
  courses: Course[]
}

export const CourseCards = ({ courses }: CourseCardsProps) => {
  const [search, setSearch] = useState('')

  const filteredCourses = useMemo(() => {
    const query = search.trim().toLowerCase()
    if (!query) return courses
    return courses.filter((course) =>
      [course.name, course.semesterTag].some((field) => field?.toLowerCase().includes(query)),
    )
  }, [courses, search])

  return (
    <div className='container mx-auto px-4 py-8'>
      <div className='relative mb-8 max-w-md'>
        <Search className='absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground' />
        <Input
          type='search'
          placeholder='Search courses...'
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          className='pl-9'
        />
      </div>

      {filteredCourses.length === 0 ? (
        <p className='text-muted-foreground'>No courses match your search.</p>
      ) : (
        <div className='grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-8 items-start justify-start'>
          <AnimatePresence>
            {filteredCourses.map((course) => (
              <motion.div
                key={course.id}
                layout
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.9 }}
                transition={{ duration: 0.2, ease: 'easeOut' }}
              >
                <CourseCard course={course} />
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      )}
    </div>
  )
}
