import { ShowForRole } from '@core/managementConsole/shared/components/ShowForRole'
import { Role } from '@tumaet/prompt-shared-state'
import { SidebarGroup, SidebarGroupContent, SidebarMenu } from '@tumaet/prompt-ui-components'
import { Activity, Archive, File, FileText, Shield, Tag, Users } from 'lucide-react'
import { InsideSidebarVisualGroup } from './components/InsideSidebarHeading'
import { InsideSidebarMenuItem } from './components/InsideSidebarMenuItem'

export const InsideGeneralSidebar = () => {
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
            <ShowForRole roles={[Role.PROMPT_LECTURER, Role.PROMPT_ADMIN]}>
              <InsideSidebarMenuItem
                icon={<File />}
                goToPath={'/management/course-templates'}
                title='Template Courses'
              />
            </ShowForRole>
            <ShowForRole
              roles={[Role.PROMPT_LECTURER, Role.PROMPT_ADMIN]}
              anyCourseRole={[Role.COURSE_LECTURER, Role.COURSE_EDITOR]}
            >
              <InsideSidebarMenuItem
                icon={<Archive />}
                goToPath={'/management/course-archive'}
                title='Archived Courses'
              />
            </ShowForRole>
          </InsideSidebarVisualGroup>
          <ShowForRole roles={[Role.PROMPT_LECTURER, Role.PROMPT_ADMIN]}>
            <InsideSidebarVisualGroup title='Students'>
              <InsideSidebarMenuItem
                icon={<Users />}
                goToPath={'/management/students'}
                title='Students'
              />
              <ShowForRole roles={[Role.PROMPT_ADMIN]}>
                <InsideSidebarMenuItem
                  icon={<Tag />}
                  goToPath={'/management/student-note-tags'}
                  title='Tags'
                />
              </ShowForRole>
            </InsideSidebarVisualGroup>
          </ShowForRole>
          <ShowForRole roles={[Role.PROMPT_ADMIN]}>
            <InsideSidebarVisualGroup title='Admin'>
              <InsideSidebarMenuItem
                icon={<Activity />}
                goToPath={'/management/system-status'}
                title='Status'
              />
              <InsideSidebarMenuItem
                icon={<Shield />}
                goToPath={'/management/admin/privacy'}
                title='Privacy'
              />
            </InsideSidebarVisualGroup>
          </ShowForRole>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarMenu>
  )
}
