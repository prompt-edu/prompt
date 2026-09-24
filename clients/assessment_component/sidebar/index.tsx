import {
  EDITOR_ROLES,
  LECTURER_ROLES,
  Role,
  type SidebarMenuItemProps,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
import { ClipboardList } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { useTutorLabel } from '../src/assessment/pages/hooks/useTutorLabel'

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

// Swaps in the phase's configured tutor display name, which a static title cannot know.
// Students never see the tutor overview entry, so they skip the config request.
export const useSidebarElement = (coursePhaseID: string): SidebarMenuItemProps => {
  const { courseId } = useParams<{ courseId: string }>()
  const isStudent = useCourseStore().isStudentOfCourse(courseId ?? '')
  const tutorLabel = useTutorLabel({ coursePhaseID, enabled: !isStudent })

  return {
    ...sidebarItems,
    subitems: sidebarItems.subitems?.map((subitem) =>
      subitem.goToPath === '/tutors' ? { ...subitem, title: `${tutorLabel} Overview` } : subitem,
    ),
  }
}

export default sidebarItems
