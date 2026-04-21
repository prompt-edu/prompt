import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@/interfaces/sidebar'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface DevOpsChallengeSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const DevOpsChallengeSidebar = React.lazy(() =>
  import('devops_challenge_component/sidebar')
    .then((module): { default: React.FC<DevOpsChallengeSidebarProps> } => ({
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
        console.warn('Failed to load challenge sidebar')
        return <DisabledSidebarMenuItem title={'DevOps Challenge Not Available'} />
      },
    })),
)
