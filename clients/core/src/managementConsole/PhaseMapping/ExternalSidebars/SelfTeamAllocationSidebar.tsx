import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface SelfTeamAllocationSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const SelfTeamAllocationSidebar = React.lazy(() =>
  import('self_team_allocation_component/sidebar')
    .then((module): { default: React.FC<SelfTeamAllocationSidebarProps> } => ({
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
        console.warn('Failed to load self team allocation sidebar')
        return <DisabledSidebarMenuItem title={'Self Team Allocation Not Available'} />
      },
    })),
)
