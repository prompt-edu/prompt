import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
} from '@tumaet/prompt-ui-components'
import SidebarHeaderComponent from './components/SidebarHeader'
import { CourseSidebarItem } from './components/CourseSidebarItem'
import { AddCourseButton } from './components/AddCourseSidebarItem'
import { useAuthStore, useCourseStore } from '@tumaet/prompt-shared-state'
import { Role } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'

export const CourseSwitchSidebar = () => {
  const { courses } = useCourseStore()
  const { courseId: openCourseId } = useParams<{ courseId: string }>()
  const coursesToShow = courses
    .filter((c) => c.id === openCourseId || (!c.template && !c.archived))
    .sort((a, b) => {
      const aIsTemporary = a.archived || a.template
      const bIsTemporary = b.archived || b.template

      if (aIsTemporary && !bIsTemporary) return -1
      if (!aIsTemporary && bIsTemporary) return 1
      return 0
    })

  const { permissions } = useAuthStore()

  const canAddCourse = permissions.some(
    (permission) => permission === Role.PROMPT_ADMIN || permission === Role.PROMPT_LECTURER,
  )

  return (
    <Sidebar
      collapsible='none'
      className='!w-[calc(var(--sidebar-width-icon)_+_1px)] min-w-[calc(var(--sidebar-width-icon)_+_1px)] border-r'
    >
      <SidebarHeaderComponent />
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent className='px-0'>
            <SidebarMenu>
              <AnimatePresence>
                {coursesToShow.map((course) => (
                  <motion.div
                    key={course.id}
                    layout
                    initial={{ opacity: 0, scale: 0.95 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.9 }}
                    transition={{ duration: 0.2, ease: 'easeOut' }}
                  >
                    <CourseSidebarItem key={course.id} course={course} />
                  </motion.div>
                ))}
              </AnimatePresence>

              {canAddCourse && <AddCourseButton />}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  )
}
