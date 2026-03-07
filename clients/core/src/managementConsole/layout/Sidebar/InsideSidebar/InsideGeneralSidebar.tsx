import { SidebarGroup, SidebarGroupContent, SidebarMenu } from '@tumaet/prompt-ui-components'
import { Archive, Activity, File, FileText, Users } from 'lucide-react'
import { InsideSidebarMenuItem } from './components/InsideSidebarMenuItem'
import { InsideSidebarVisualGroup } from './components/InsideSidebarHeading'
import { useAuthStore, Role } from '@tumaet/prompt-shared-state'

export const InsideGeneralSidebar = () => {
  const { permissions } = useAuthStore()
  const isPromptAdmin = permissions.includes(Role.PROMPT_ADMIN)

  return (
    <SidebarMenu>
      <SidebarGroup>
        <SidebarGroupContent className='flex flex-col gap-5'>
          <InsideSidebarVisualGroup title='Courses'>
            <InsideSidebarMenuItem
              icon={<FileText />}
              goToPath={'/management/courses'}
              title='Courses'
            />
            <InsideSidebarMenuItem
              icon={<File />}
              goToPath={'/management/course_templates'}
              title='Template Courses'
            />
            <InsideSidebarMenuItem
              icon={<Archive />}
              goToPath={'/management/course_archive'}
              title='Archived Courses'
            />
          </InsideSidebarVisualGroup>
          <InsideSidebarVisualGroup title='Students'>
            <InsideSidebarMenuItem
              icon={<Users />}
              goToPath={'/management/students'}
              title='Students'
            />
          </InsideSidebarVisualGroup>
          {isPromptAdmin && (
            <InsideSidebarVisualGroup title='System'>
              <InsideSidebarMenuItem
                icon={<Activity />}
                goToPath={'/management/system-status'}
                title='System Status'
              />
            </InsideSidebarVisualGroup>
          )}
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarMenu>
  )
}
