import { cn, Sidebar, useSidebar } from '@tumaet/prompt-ui-components'
import type * as React from 'react'
import { CourseSwitchSidebar } from './CourseSwitchSidebar/CourseSwitchSidebar'
import { InsideSidebar } from './InsideSidebar/InsideSidebar'

type AppSidebarProps = React.ComponentProps<typeof Sidebar>

function AppSidebarContent() {
  const { isMobile, state } = useSidebar()
  const showInsideSidebar = isMobile || state === 'expanded'

  return (
    <>
      <CourseSwitchSidebar />
      {showInsideSidebar && <InsideSidebar />}
    </>
  )
}

export function AppSidebar({ className, ...props }: AppSidebarProps) {
  return (
    <Sidebar
      collapsible='icon'
      className={cn('overflow-hidden [&>[data-sidebar=sidebar]]:flex-row print:hidden', className)}
      {...props}
    >
      <AppSidebarContent />
    </Sidebar>
  )
}
