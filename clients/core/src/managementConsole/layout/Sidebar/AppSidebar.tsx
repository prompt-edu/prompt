import * as React from 'react'

import { Sidebar } from '@tumaet/prompt-ui-components'
import { InsideSidebar } from './InsideSidebar/InsideSidebar'
import { CourseSwitchSidebar } from './CourseSwitchSidebar/CourseSwitchSidebar'

type AppSidebarProps = React.ComponentProps<typeof Sidebar>

export function AppSidebar({ ...props }: AppSidebarProps) {
  return (
    <Sidebar
      collapsible='icon'
      className='overflow-hidden *:data-[sidebar=sidebar]:flex-row'
      {...props}
    >
      {/* This is the first sidebar */}
      <CourseSwitchSidebar />
      <InsideSidebar />
    </Sidebar>
  )
}
