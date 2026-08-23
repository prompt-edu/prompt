import { EDITOR_ROLES, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Mail } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'

export const CourseMailingSidebar = ({ rootPath, title }: { rootPath: string; title: string }) => {
  const courseMailingSidebarItem: SidebarMenuItemProps = {
    title: 'Mailing',
    icon: <Mail />,
    goToPath: '/mailing',
    requiredPermissions: EDITOR_ROLES,
  }
  return (
    <ExternalSidebarComponent
      title={title}
      rootPath={rootPath}
      sidebarElement={courseMailingSidebarItem}
    />
  )
}
