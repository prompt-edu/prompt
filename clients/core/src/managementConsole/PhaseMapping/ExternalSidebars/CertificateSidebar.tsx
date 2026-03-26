import React from 'react'
import { DisabledSidebarMenuItem } from '../../layout/Sidebar/InsideSidebar/components/DisabledSidebarMenuItem'
import { SidebarMenuItemProps } from '@/interfaces/sidebar'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface CertificateSidebarProps {
  rootPath: string
  title?: string
  coursePhaseID: string
}

export const CertificateSidebar = React.lazy(() =>
  import('certificate_component/sidebar')
    .then((module): { default: React.FC<CertificateSidebarProps> } => ({
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
        console.warn('Failed to load certificate sidebar')
        return <DisabledSidebarMenuItem title={'Certificate Not Available'} />
      },
    })),
)
