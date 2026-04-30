import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface InfrastructureSetupSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const InfrastructureSetupSidebar = React.lazy(() =>
  import('infrastructure_setup_component/sidebar')
    .then((module): { default: React.FC<InfrastructureSetupSidebarProps> } => ({
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
        console.warn('Failed to load infrastructure setup sidebar')
        return <DisabledSidebarMenuItem title={'Infrastructure Setup Not Available'} />
      },
    })),
)
