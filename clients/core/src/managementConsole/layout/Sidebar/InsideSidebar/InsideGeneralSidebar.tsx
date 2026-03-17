import { SidebarGroup, SidebarGroupContent, SidebarMenu } from '@tumaet/prompt-ui-components'
import { Archive, File, FileText, Tag, Users } from 'lucide-react'
import { InsideSidebarMenuItem } from './components/InsideSidebarMenuItem'
import { InsideSidebarVisualGroup } from './components/InsideSidebarHeading'
import { Role, getPermissionString, useAuthStore } from '@tumaet/prompt-shared-state'

export const InsideGeneralSidebar = () => {
  const { permissions } = useAuthStore()
  const isPromptAdmin = permissions.includes(getPermissionString(Role.PROMPT_ADMIN))

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
            {isPromptAdmin && (
              <InsideSidebarMenuItem
                icon={<Tag />}
                goToPath={'/management/student-note-tags'}
                title='Tags'
              />
            )}
          </InsideSidebarVisualGroup>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarMenu>
  )
}
