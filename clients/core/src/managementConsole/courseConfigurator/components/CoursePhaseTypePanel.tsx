import { useState } from 'react'
import { ChevronRight } from 'lucide-react'
import { motion } from 'framer-motion'
import { useCourseConfigurationState } from '../zustand/useCourseConfigurationStore'
import { CoursePhaseType } from '../interfaces/coursePhaseType'
import { CoursePhaseTypePanelItem } from './CoursePhaseTypePanelItem'

interface CoursePhaseTypePanelProps {
  canEdit: boolean
}

export const CoursePhaseTypePanel = ({ canEdit }: CoursePhaseTypePanelProps) => {
  const { coursePhaseTypes, coursePhases } = useCourseConfigurationState()
  const [isExpanded, setIsExpanded] = useState(false)

  const courseHasInitialPhase = coursePhases.some((phase) => phase.isInitialPhase)

  const coursePhaseTypesOrdered = [...coursePhaseTypes].sort((a, b) => {
    if (a.initialPhase && !b.initialPhase) return -1
    if (!a.initialPhase && b.initialPhase) return 1
    return 0
  })

  const isDraggable = (phase: CoursePhaseType) => {
    if (!canEdit) return false
    if (phase.initialPhase && courseHasInitialPhase) return false
    if (!courseHasInitialPhase && !phase.initialPhase) return false
    return true
  }

  return (
    <motion.div
      className='absolute left-0 top-0 bottom-0 z-50 flex flex-col border-r bg-background overflow-hidden'
      animate={{ width: isExpanded ? 220 : 40 }}
      transition={{ duration: 0.2, ease: 'easeInOut' }}
      onMouseEnter={() => setIsExpanded(true)}
      onMouseLeave={() => setIsExpanded(false)}
    >
      {/* Collapsed indicator */}
      <motion.div
        className='absolute inset-0 flex flex-col items-center pt-4 gap-2 pointer-events-none text-muted-foreground'
        animate={{ opacity: isExpanded ? 0 : 1 }}
        transition={{ duration: 0.1 }}
      >
        <ChevronRight className='h-4 w-4' />
        <span className='transform -rotate-90 text-nowrap mt-12 -translate-x-[2px]'>
          Course phases
        </span>
      </motion.div>

      {/* Expanded content */}
      <motion.div
        className='flex flex-col h-full min-w-[220px]'
        animate={{ opacity: isExpanded ? 1 : 0 }}
        transition={{ duration: 0.15 }}
      >
        <div className='px-3 py-3 border-b'>
          <h2 className='text-sm font-semibold whitespace-nowrap'>Course Phases</h2>
        </div>
        <div className='flex-1 overflow-y-auto p-3 flex flex-col gap-2'>
          {coursePhaseTypesOrdered.map((phase) => (
            <CoursePhaseTypePanelItem
              key={phase.id}
              phase={phase}
              isDraggable={isDraggable(phase)}
              onDragStart={() => setIsExpanded(false)}
            />
          ))}
        </div>
      </motion.div>
    </motion.div>
  )
}
