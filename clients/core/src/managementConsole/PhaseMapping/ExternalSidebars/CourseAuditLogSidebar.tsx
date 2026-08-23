import { LECTURER_ROLES, type SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ScrollText } from 'lucide-react'
import { ExternalSidebarComponent } from './ExternalSidebar'

export const CourseAuditLogSidebar = ({ rootPath, title }: { rootPath: string; title: string }) => {
  const auditLogSidebarItem: SidebarMenuItemProps = {
    title: 'Audit Log',
    icon: <ScrollText />,
    goToPath: '/audit-log',
    requiredPermissions: LECTURER_ROLES,
  }
  return (
    <ExternalSidebarComponent
      title={title}
      rootPath={rootPath}
      sidebarElement={auditLogSidebarItem}
    />
  )
}
