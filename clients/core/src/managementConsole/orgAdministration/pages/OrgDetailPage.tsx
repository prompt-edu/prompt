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
    <QueryGate queries={[orgQuery]}>
      {() => {
        const org = orgQuery.data
        if (!org) {
          return null
        }
        return (
          <div className='w-full space-y-6'>
            <div>
              <ManagementPageHeader>{org.name}</ManagementPageHeader>
              <p className='-mt-4 text-muted-foreground'>{org.slug}</p>
            </div>
            <OrgDetailsSettings key={`details-${org.id}`} org={org} />
            <QueryGate queries={[orgsQuery]} loadingFallback={null}>
              {() =>
                orgsQuery.data && (
                  <OrgHierarchySettings
                    key={`hierarchy-${org.id}-${org.parentOrgID ?? 'none'}`}
                    org={org}
                    orgs={orgsQuery.data}
                  />
                )
              }
            </QueryGate>
            <OrgDangerZone org={org} />
          </div>
        )
      }}
    </QueryGate>
  )
}
