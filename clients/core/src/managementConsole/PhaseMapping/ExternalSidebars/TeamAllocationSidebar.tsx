import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface TeamAllocationSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const TeamAllocationSidebar = React.lazy(() =>
  import('team_allocation_component/sidebar')
    .then((module): { default: React.FC<TeamAllocationSidebarProps> } => ({
      default: ({ title, rootPath, coursePhaseID }) => {
        const sidebarElement: SidebarMenuItemProps = module.default || {}
        return (
          <ExternalSidebarComponent
            title={title}
            rootPath={rootPath}
            sidebarElement={sidebarElement}
            coursePhaseID={coursePhaseID}
          />
        )
      },
    }))
    .catch((): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load team allocation sidebar')
        return <DisabledSidebarMenuItem title={'Team Allocation Not Available'} />
      },
    })),
)
