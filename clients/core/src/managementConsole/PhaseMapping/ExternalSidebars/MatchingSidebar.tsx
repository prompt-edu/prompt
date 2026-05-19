import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface MatchingSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const MatchingSidebar = React.lazy(() =>
  import('matching_component/sidebar')
    .then((module): { default: React.FC<MatchingSidebarProps> } => ({
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
        console.warn('Failed to load matching sidebar')
        return <DisabledSidebarMenuItem title={'Matching Not Available'} />
      },
    })),
)
