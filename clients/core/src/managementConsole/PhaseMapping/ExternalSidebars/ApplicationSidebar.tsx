import { Role } from '@tumaet/prompt-shared-state'
import { FileUser } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'

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
    requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    subitems: [
      {
        title: 'Participants',
        goToPath: '/participants',
        requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
      },
      {
        title: 'Questions',
        goToPath: '/questions',
        requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
      },
      {
        title: 'Mailing',
        goToPath: '/mailing',
        requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
      },
      {
        title: 'Settings',
        goToPath: '/settings',
        requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
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
