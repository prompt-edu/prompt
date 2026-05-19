import { Sidebar, SidebarContent, SidebarHeader } from '@tumaet/prompt-ui-components'
import { useLocation } from 'react-router-dom'
import { InsideCourseSidebar } from './InsideCourseSidebar'
import { InsideGeneralSidebar } from './InsideGeneralSidebar'
import { PromptLogo } from './components/PromptLogo'

export const InsideSidebar = () => {
  // set the correct header
  const location = useLocation()
  const isCourseSidebar = location.pathname.startsWith('/management/course/')

  return (
    <Sidebar collapsible='none' className='flex w-sidebar max-w-sidebar'>
      <SidebarHeader className='flex h-14 border-b justify-center items-center'>
        <PromptLogo />
      </SidebarHeader>
      <SidebarContent>
        {isCourseSidebar ? <InsideCourseSidebar /> : <InsideGeneralSidebar />}
      </SidebarContent>
    </Sidebar>
  )
}
