import { useAIEnabled } from '@core/network/hooks/useAIEnabled'
import type { SidebarMenuItemProps } from '@tumaet/prompt-shared-state'
import React, { Suspense } from 'react'
import { loadAIRemote } from '../aiRemote'
import { ExternalSidebarComponent } from './ExternalSidebar'

interface AISidebarProps {
  rootPath: string
  title?: string
}

const RemoteAISidebar = React.lazy(() =>
  loadAIRemote<typeof import('ai_component/sidebar')>('sidebar')
    .then((module): { default: React.FC<AISidebarProps> } => ({
      default: ({ title, rootPath }) => {
        const sidebarElement: SidebarMenuItemProps = module.default || {}
        return (
          <ExternalSidebarComponent
            title={title}
            rootPath={rootPath}
            sidebarElement={sidebarElement}
          />
        )
      },
    }))
    .catch((): { default: React.FC } => ({
      default: () => {
        console.warn('Failed to load AI sidebar')
        return null
      },
    })),
)

// The AI remote is only deployed with the AI server, so it is not even registered while AI is off.
export const AISidebar = ({ rootPath, title }: AISidebarProps) => {
  const aiEnabled = useAIEnabled()

  if (!aiEnabled) {
    return null
  }

  return (
    <Suspense fallback={null}>
      <RemoteAISidebar rootPath={`${rootPath}/ai`} title={title} />
    </Suspense>
  )
}
