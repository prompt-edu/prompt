import type { Org, OrgRole } from '@core/interfaces/org'
import { isNotFoundError, useOrg, useOrgs } from '@core/network/hooks/useOrgs'
import { ManagementPageHeader, QueryGate } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'
import { EmptyPage } from '../../shared/components/EmptyPage'
import { OrgDangerZone } from '../components/OrgDangerZone'
import { OrgDetailsSettings } from '../components/OrgDetailsSettings'
import { OrgHierarchySettings } from '../components/OrgHierarchySettings'
import { OrgMembersSettings } from '../components/OrgMembersSettings'

// Who looks at the org: a PROMPT admin from the admin console, or a member of the org from its own
// entry. Nesting grants nothing, so the hierarchy and the deletion stay with PROMPT admins.
export type OrgViewer = 'PROMPT_Admin' | OrgRole

interface OrgDetailPageProps {
  viewer: OrgViewer
}

const PromptAdminOrgSettings = ({ org }: { org: Org }) => {
  const orgsQuery = useOrgs()

  return (
    <>
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
    </>
  )
}

export const OrgDetailPage = ({ viewer }: OrgDetailPageProps) => {
  const { orgId } = useParams<{ orgId: string }>()
  const orgQuery = useOrg(orgId)

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
            <OrgDetailsSettings
              key={`details-${org.id}`}
              org={org}
              readOnly={viewer === 'Member'}
            />
            {viewer !== 'Member' && <OrgMembersSettings org={org} />}
            {viewer === 'PROMPT_Admin' && <PromptAdminOrgSettings org={org} />}
          </div>
        )
      }}
    </QueryGate>
  )
}
