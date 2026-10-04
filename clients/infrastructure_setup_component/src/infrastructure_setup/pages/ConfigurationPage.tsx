import { ManagementPageHeader, Separator } from '@tumaet/prompt-ui-components'
import { useEffect } from 'react'
import { useLocation, useParams } from 'react-router-dom'
import { ProvidersSection } from '../components/configuration/ProvidersSection'
import { ResourcesSection } from '../components/configuration/ResourcesSection'
import { SemesterTagSection } from '../components/configuration/SemesterTagSection'

export const ConfigurationPage = () => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const { hash } = useLocation()

  // The provisioning checklist links to one section; the router does not scroll to an
  // anchor on its own.
  useEffect(() => {
    if (hash) {
      document.getElementById(hash.slice(1))?.scrollIntoView()
    }
  }, [hash])

  if (!courseId || !phaseId) return null

  return (
    <div className='max-w-5xl space-y-8'>
      <div>
        <ManagementPageHeader>Configuration</ManagementPageHeader>
        <p className='text-muted-foreground'>
          Set up what this phase provisions. Nothing is created until you start it on the
          Provisioning page.
        </p>
      </div>
      <SemesterTagSection courseId={courseId} coursePhaseID={phaseId} />
      <Separator />
      <ProvidersSection coursePhaseID={phaseId} />
      <Separator />
      <ResourcesSection coursePhaseID={phaseId} />
    </div>
  )
}

export default ConfigurationPage
