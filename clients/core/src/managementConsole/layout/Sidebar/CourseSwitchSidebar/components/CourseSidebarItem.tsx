import type { Course } from '@tumaet/prompt-shared-state'
import { SidebarMenuButton, SidebarMenuItem, useSidebar } from '@tumaet/prompt-ui-components'
import { useNavigate, useParams } from 'react-router-dom'
import { CourseAvatar } from './CourseAvatar'
import { CourseSidebarItemTooltip } from './CourseSidebarItemTooltip'

interface CourseSidebarItemProps {
  course: Course
}

export const CourseSidebarItem = ({ course }: CourseSidebarItemProps) => {
  const { setOpen } = useSidebar()
  const navigate = useNavigate()
  const { courseId } = useParams<{ courseId: string }>()

  const isActive = course.id === courseId
  const bgColor = course.studentReadableData?.['bg-color'] || 'bg-gray-100'
  const iconName = course.studentReadableData?.['icon'] || 'graduation-cap'

  const containerRing = isActive
    ? course.template
      ? 'after:absolute after:inset-0 after:rounded-lg after:border-2 after:border-dashed after:border-black'
      : course.archived
        ? 'after:absolute after:inset-0 after:rounded-lg after:border-2 after:border-muted-foreground'
        : 'after:absolute after:inset-0 after:rounded-lg after:border-2 after:border-primary'
    : ''

  return (
    <SidebarMenuItem key={course.id}>
      <SidebarMenuButton
        size='lg'
        tooltip={{
          children: <CourseSidebarItemTooltip course={course} />,
          hidden: false,
        }}
        onClick={() => {
          setOpen(true)
          navigate(`/management/course/${course.id}`)
        }}
        isActive={isActive}
        className='min-w-12 min-h-12 p-0'
      >
        <div
          className={`relative flex aspect-square size-12 items-center justify-center ${containerRing} ${course.archived ? 'opacity-60' : ''}`}
        >
          <CourseAvatar bgColor={bgColor} iconName={iconName} isActive={isActive} />
        </div>
      </SidebarMenuButton>
    </SidebarMenuItem>
  )
}
