import {
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@tumaet/prompt-ui-components'
import { House } from 'lucide-react'
import React from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

const SidebarHeaderComponent = () => {
  const { setOpen } = useSidebar()

  const location = useLocation()
  const navigate = useNavigate()

  const isActive = location.pathname.startsWith('/management/courses')

  return (
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            size='lg'
            asChild
            className='min-w-12 min-h-12 p-0'
            tooltip={{
              children: 'Home',
              hidden: false,
            }}
            onClick={() => {
              setOpen(true)
              navigate('/management/courses')
            }}
            isActive={isActive}
          >
            <div
              className={`relative flex aspect-square size-12 items-center justify-center ${
                isActive
                  ? 'after:absolute after:inset-0 after:rounded-lg after:border-2 after:border-secondary'
                  : ''
              }`}
            >
              <div
                className={`
                  flex aspect-square items-center justify-center rounded-lg bg-sidebar-secondary text-sidebar-primary-foreground 
                  ${isActive ? 'size-12' : 'size-10'}
                  `}
              >
                <House className='text-black dark:text-white' />
              </div>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
  )
}

export default SidebarHeaderComponent
