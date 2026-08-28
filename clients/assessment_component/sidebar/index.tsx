import {
  EDITOR_ROLES,
  LECTURER_ROLES,
  Role,
  type SidebarMenuItemProps,
} from '@tumaet/prompt-shared-state'
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
      requiredPermissions: EDITOR_ROLES,
    },
    {
      title: 'Tutor Overview',
      goToPath: '/tutors',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Self Evaluations',
      goToPath: '/self-evaluations',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Peer Evaluations',
      goToPath: '/peer-evaluations',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Statistics',
      goToPath: '/statistics',
      requiredPermissions: LECTURER_ROLES,
    },
    {
      title: 'Settings',
      goToPath: '/settings',
      requiredPermissions: LECTURER_ROLES,
    },
  ],
}

export default sidebarItems
