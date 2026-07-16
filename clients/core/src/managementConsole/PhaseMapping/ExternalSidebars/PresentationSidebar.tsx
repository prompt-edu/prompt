import type { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface PresentationSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const PresentationSidebar = React.lazy(() =>
  import('presentation_component/sidebar')
    .then((module): { default: React.FC<PresentationSidebarProps> } => ({
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
        console.warn('Failed to load presentation sidebar')
        return <DisabledSidebarMenuItem title={'Presentation Not Available'} />
      },
    })),
)
