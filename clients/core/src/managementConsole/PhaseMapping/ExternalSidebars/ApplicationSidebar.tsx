import { LECTURER_ROLES, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { FileUser } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'

export const ApplicationSidebar = ({
  rootPath,
  title,
  coursePhaseID,
}: {
  rootPath: string
  title: string
  coursePhaseID: string
}) => {
  const applicationSidebarItems: SidebarMenuItemProps = {
    title: 'Application',
    icon: <FileUser />,
    goToPath: '',
    requiredPermissions: LECTURER_ROLES,
    subitems: [
      {
        title: 'Participants',
        goToPath: '/participants',
        requiredPermissions: LECTURER_ROLES,
      },
      {
        title: 'Questions',
        goToPath: '/questions',
        requiredPermissions: LECTURER_ROLES,
      },
      {
        title: 'Settings',
        goToPath: '/settings',
        requiredPermissions: LECTURER_ROLES,
      },
    ],
  }
  return (
    <ExternalSidebarComponent
      title={title}
      rootPath={rootPath}
      sidebarElement={applicationSidebarItems}
      coursePhaseID={coursePhaseID}
    />
  )
}
