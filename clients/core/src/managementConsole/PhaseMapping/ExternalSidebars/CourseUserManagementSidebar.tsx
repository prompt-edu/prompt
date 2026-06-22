import { Role } from '@tumaet/prompt-shared-state'
import { Users } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'

export const CourseUserManagementSidebar = ({
  rootPath,
  title,
}: {
  rootPath: string
  title: string
}) => {
  const item: SidebarMenuItemProps = {
    title: 'User Management',
    icon: <Users />,
    goToPath: '/user-management',
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
  }
  return <ExternalSidebarComponent title={title} rootPath={rootPath} sidebarElement={item} />
}
