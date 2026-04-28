import { Role } from '@tumaet/prompt-shared-state'
import { Route } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'

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
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_EDITOR],
  }
  return (
    <ExternalSidebarComponent
      title={title}
      rootPath={rootPath}
      sidebarElement={courseConfiguratorSidebarItems}
    />
  )
}
