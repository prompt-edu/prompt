import { LECTURER_ROLES, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Users } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface CourseUserManagementSidebarProps {
  rootPath: string
  title: string
}

export const CourseUserManagementSidebar = ({
  rootPath,
  title,
}: CourseUserManagementSidebarProps) => {
  const item: SidebarMenuItemProps = {
    title: 'User Management',
    icon: <Users />,
    goToPath: '/user-management',
    requiredPermissions: LECTURER_ROLES,
  }
  return <ExternalSidebarComponent title={title} rootPath={rootPath} sidebarElement={item} />
}
