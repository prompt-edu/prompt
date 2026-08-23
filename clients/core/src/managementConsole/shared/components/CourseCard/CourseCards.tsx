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
  const query = search.trim().toLowerCase()

  const filteredCourses = useMemo(() => {
    if (!query) return courses
    return courses.filter((course) =>
      [course.name, course.semesterTag].some((field) => field?.toLowerCase().includes(query)),
    )
  }, [courses, query])

  return (
    <div className='flex flex-col gap-6'>
      <div className='relative max-w-md'>
        <Search className='pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
        {/* Remotes ship their own Tailwind build, so an unprefixed pl-9 loses to their .px-3 */}
        <Input
          type='search'
          aria-label='Search courses'
          placeholder='Search courses...'
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          className='pl-9!'
        />
      </div>

      <div className='grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-8 items-start justify-start'>
        <AnimatePresence>
          {/* Only a search with no matches is empty here; callers guard the no-courses-at-all case. */}
          {query && filteredCourses.length === 0 && (
            <motion.p
              key='empty'
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className='text-muted-foreground col-span-full'
            >
              No courses match your search.
            </motion.p>
          )}
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
    </div>
  )
}
