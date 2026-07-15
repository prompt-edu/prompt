import type { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface ExampleSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const ExampleSidebar = React.lazy(() =>
  import('example_component/sidebar')
    .then((module): { default: React.FC<ExampleSidebarProps> } => ({
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
        console.warn('Failed to load example routes')
        return <DisabledSidebarMenuItem title={'Example Not Available'} />
      },
    })),
)
