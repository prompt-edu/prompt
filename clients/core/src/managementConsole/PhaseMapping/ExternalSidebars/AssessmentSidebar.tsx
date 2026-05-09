import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface TemplateSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const AssessmentSidebar = React.lazy(() =>
  import('assessment_component/sidebar')
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
        console.warn('Failed to load assessment routes')
        return <DisabledSidebarMenuItem title={'Assessment Not Available'} />
      },
    })),
)
