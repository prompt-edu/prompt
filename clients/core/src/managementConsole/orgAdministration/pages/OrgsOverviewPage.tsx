import type { CreateOrg } from '@core/interfaces/org'
import { useCreateOrg, useOrgs } from '@core/network/hooks/useOrgs'
import { Button, ManagementPageHeader, PromptTable, QueryGate } from '@tumaet/prompt-ui-components'
import { Plus } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { CreateOrgDialog } from '../components/CreateOrgDialog'
import { orgTableColumns } from '../components/orgTableColumns'
import { type OrgRow, withParentLabels } from '../utils/orgHierarchy'

export const OrgsOverviewPage = () => {
  const navigate = useNavigate()
  const orgsQuery = useOrgs()
  const createOrg = useCreateOrg()
  const [createOpen, setCreateOpen] = useState(false)

  const rows = useMemo(() => withParentLabels(orgsQuery.data ?? []), [orgsQuery.data])

  const handleCreate = (org: CreateOrg) => {
    createOrg.mutate(org, { onSuccess: () => setCreateOpen(false) })
  }

  return (
    <div className='w-full'>
      <div className='flex items-center justify-between'>
        <ManagementPageHeader>Orgs</ManagementPageHeader>
        <Button onClick={() => setCreateOpen(true)} disabled={!orgsQuery.isSuccess}>
          <Plus className='h-4 w-4 mr-2' />
          Add Org
        </Button>
      </div>
      <p className='text-muted-foreground mb-6'>
        Orgs are organizational units such as university chairs, and they can be nested. Each org
        has an Admin and a Member role in Keycloak.
      </p>

      <QueryGate queries={[orgsQuery]}>
        <PromptTable<OrgRow>
          data={rows}
          columns={orgTableColumns}
          onRowClick={(org) => navigate(`/management/admin/orgs/${org.id}`)}
        />
      </QueryGate>

      {createOpen && (
        <CreateOrgDialog
          orgs={orgsQuery.data ?? []}
          isPending={createOrg.isPending}
          onClose={() => setCreateOpen(false)}
          onSubmit={handleCreate}
        />
      )}
    </div>
  )
}
