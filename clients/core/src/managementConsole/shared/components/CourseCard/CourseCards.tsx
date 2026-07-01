import { Course } from '@tumaet/prompt-shared-state'
import { AnimatePresence, motion } from 'framer-motion'
import { CourseCard } from './CourseCard'

interface CourseCardsProps {
  courses: Course[]
}

export const CourseCards = ({ courses }: CourseCardsProps) => {
  return (
    <div className='container mx-auto px-4 py-8 grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-8 items-start justify-start'>
      <AnimatePresence>
        {courses.map((course) => (
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
  )
}
