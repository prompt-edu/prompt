import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Mail } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'

export const CourseMailingSidebar = ({ rootPath, title }: { rootPath: string; title: string }) => {
  const courseMailingSidebarItem: SidebarMenuItemProps = {
    title: 'Mailing',
    icon: <Mail />,
    goToPath: '/mailing',
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_EDITOR],
  }
  return (
    <ExternalSidebarComponent
      title={title}
      rootPath={rootPath}
      sidebarElement={courseMailingSidebarItem}
    />
  )
}
