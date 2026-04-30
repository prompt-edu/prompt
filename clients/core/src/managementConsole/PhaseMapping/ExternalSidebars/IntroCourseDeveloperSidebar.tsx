import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface IntroCourseDeveloperSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const IntroCourseDeveloperSidebar = React.lazy(() =>
  import('intro_course_developer_component/sidebar')
    .then((module): { default: React.FC<IntroCourseDeveloperSidebarProps> } => ({
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
        console.warn('Failed to load intro course developer sidebar')
        return <DisabledSidebarMenuItem title={'Intro Course Not Available'} />
      },
    })),
)
