import '../src/loadStyles'

import { Role, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ClipboardList } from 'lucide-react'

const sidebarItems: SidebarMenuItemProps = {
  title: 'Assessment Component',
  icon: <ClipboardList />,
  goToPath: '',
  requiredPermissions: [
    Role.PROMPT_ADMIN,
    Role.COURSE_LECTURER,
    Role.COURSE_EDITOR,
    Role.COURSE_STUDENT,
  ],
  subitems: [
    {
      title: 'Participants',
      goToPath: '/participants',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER, Role.COURSE_EDITOR],
    },
    {
      title: 'Tutor Overview',
      goToPath: '/tutors',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Self Evaluations',
      goToPath: '/self-evaluations',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Peer Evaluations',
      goToPath: '/peer-evaluations',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Statistics',
      goToPath: '/statistics',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
    {
      title: 'Settings',
      goToPath: '/settings',
      requiredPermissions: [Role.PROMPT_ADMIN, Role.COURSE_LECTURER],
    },
  ],
}

export default sidebarItems
