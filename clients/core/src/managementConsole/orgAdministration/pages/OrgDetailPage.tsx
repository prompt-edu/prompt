import { isNotFoundError, useOrg, useOrgs } from '@core/network/hooks/useOrgs'
import { ManagementPageHeader, QueryGate } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'
import { EmptyPage } from '../../shared/components/EmptyPage'
import { OrgDangerZone } from '../components/OrgDangerZone'
import { OrgDetailsSettings } from '../components/OrgDetailsSettings'
import { OrgHierarchySettings } from '../components/OrgHierarchySettings'

export const OrgDetailPage = () => {
  const { orgId } = useParams<{ orgId: string }>()
  const orgQuery = useOrg(orgId)
  const orgsQuery = useOrgs()

  if (isNotFoundError(orgQuery.error)) {
    return <EmptyPage message='Org not found' />
  }

  return (
    <QueryGate queries={[orgQuery, orgsQuery]}>
      {() => {
        const org = orgQuery.data
        const orgs = orgsQuery.data
        if (!org || !orgs) {
          return null
        }
        return (
          <div className='w-full space-y-6'>
            <div>
              <ManagementPageHeader>{org.name}</ManagementPageHeader>
              <p className='-mt-4 text-muted-foreground'>{org.slug}</p>
            </div>
            <OrgDetailsSettings key={org.id} org={org} />
            <OrgHierarchySettings key={org.id} org={org} orgs={orgs} />
            <OrgDangerZone org={org} />
          </div>
        )
      }}
    </QueryGate>
  )
}
