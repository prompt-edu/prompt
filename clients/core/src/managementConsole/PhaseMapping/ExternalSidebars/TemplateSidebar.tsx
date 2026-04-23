import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface TemplateSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const TemplateSidebar = React.lazy(() =>
  import('template_component/sidebar')
    .then((module): { default: React.FC<TemplateSidebarProps> } => ({
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
        console.warn('Failed to load template routes')
        return <DisabledSidebarMenuItem title={'Template Not Available'} />
      },
    })),
)
