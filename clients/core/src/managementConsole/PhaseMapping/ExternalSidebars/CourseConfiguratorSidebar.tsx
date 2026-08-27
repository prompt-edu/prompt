import { EDITOR_ROLES, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { Route } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'

export const CourseConfiguratorSidebar = ({
  rootPath,
  title,
}: {
  rootPath: string
  title: string
}) => {
  const courseConfiguratorSidebarItems: SidebarMenuItemProps = {
    title: 'Configure Course Phases',
    icon: <Route />,
    goToPath: '/configurator',
    requiredPermissions: EDITOR_ROLES,
  }
  return (
    <ExternalSidebarComponent
      title={title}
      rootPath={rootPath}
      sidebarElement={courseConfiguratorSidebarItems}
    />
  )
}
